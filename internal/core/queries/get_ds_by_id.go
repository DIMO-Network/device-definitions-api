//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
)

type GetDeviceStyleByIDQuery struct {
	DeviceStyleID string `json:"device_style_id"`
}

func (*GetDeviceStyleByIDQuery) Key() string { return "GetDeviceStyleByIDQuery" }

type GetDeviceStyleByIDQueryHandler struct {
	DBS        func() *db.ReaderWriter
	catalogSvc gateways.DeviceDefinitionCatalogService
}

func NewGetDeviceStyleByIDQueryHandler(dbs func() *db.ReaderWriter, onchainSvc gateways.DeviceDefinitionCatalogService) GetDeviceStyleByIDQueryHandler {
	return GetDeviceStyleByIDQueryHandler{DBS: dbs, catalogSvc: onchainSvc}
}

func (ch GetDeviceStyleByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceStyleByIDQuery)

	ds, err := models.DeviceStyles(models.DeviceStyleWhere.ID.EQ(qry.DeviceStyleID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device style id: %s", qry.DeviceStyleID),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device styles"),
		}
	}
	dd, _, err := ch.catalogSvc.GetTemplateByID(ctx, ds.DefinitionID)
	if err != nil {
		// GetTemplateByID's error is a genuine "does not exist" only when it
		// is ErrTemplateNotFound (checked by identity, never by message): a
		// catalog outage -- a 5xx, a timeout, a decode failure -- must not be
		// reclassified as not-found, or a caller reacting to the spurious 404
		// could create a duplicate definition for a vehicle that already
		// exists, during the worst possible moment to do so.
		if errors.Is(err, gateways.ErrTemplateNotFound) {
			return nil, &exceptions.NotFoundError{Err: fmt.Errorf("device definition not found in catalog: %s: %w", ds.DefinitionID, err)}
		}
		return nil, &exceptions.InternalError{Err: fmt.Errorf("failed to get device definition %s from catalog: %w", ds.DefinitionID, err)}
	}

	deviceStyleResult := coremodels.GetDeviceStyleQueryResult{
		ID:                 ds.ID,
		DefinitionID:       ds.DefinitionID,
		Name:               ds.Name,
		ExternalStyleID:    ds.ExternalStyleID,
		Source:             ds.Source,
		SubModel:           ds.SubModel,
		HardwareTemplateID: ds.HardwareTemplateID.String,
		DeviceDefinition: coremodels.GetDeviceDefinitionStyleQueryResult{
			// The style's name is the signal the trim matcher consumes, so
			// the definition served here is the trim this style actually is,
			// not the shared part of the template it belongs to.
			DeviceAttributes: styleDeviceAttributes(dd, ds.Name, ds.Metadata.JSON),
		},
	}

	return deviceStyleResult, nil
}

// styleDeviceAttributes serves the attributes of the trim this style is.
//
// A template carries at the top level only the attributes every trim agrees
// on -- by design, so that nothing which varies by trim is asserted for all of
// them. Copying those alone dropped every trim-varying attribute (mpg, msrp,
// seating, fuel tank) and left a caller unable to tell an attribute the
// template does not have from one living on a trim it was not given. For a
// model sold as combustion, hybrid and plug-in hybrid, nothing carried a
// powertrain at all: the name heuristic misses names like "Prime SE", and the
// response asserted internal combustion for a plug-in hybrid.
//
// The style's name is exactly services.MatchTrim's styleName signal, so the
// trim is resolved here. An exact match serves that trim's attributes over the
// template's; an ambiguous one serves only what every candidate agrees on; no
// match serves the template's own. In every case what comes back is what some
// trim actually has, never a merge across trims.
func styleDeviceAttributes(tmpl *coremodels.Template, styleName string, styleMetadata []byte) []coremodels.DeviceTypeAttributeEditor {
	resolved := services.MatchTrim(tmpl, services.MatchSignals{StyleName: styleName})
	attrs := convertTemplateAttributesToDeviceAttributes(resolved.Attributes)

	powerTrainType := powertrainForStyle(resolved, styleName, styleMetadata)
	for i, item := range attrs {
		if item.Name == common.PowerTrainType {
			attrs[i].Value = powerTrainType
			return attrs
		}
	}
	return append(attrs, coremodels.DeviceTypeAttributeEditor{
		Name:        common.PowerTrainType,
		Label:       common.PowerTrainType,
		Description: common.PowerTrainType,
		Type:        common.DefaultDeviceType,
		Value:       powerTrainType,
	})
}

// powertrainForStyle decides the one powertrain this style is reported with.
//
// In order: what the style row itself recorded at decode time, which is the
// only value observed for this style rather than inferred; then an exactly
// matched trim's, which is more specific than any reading of the name; then
// the name heuristic, which is how a style that narrowed to no single trim has
// always been answered; then whatever the resolution left -- the template's
// own value, or the one every ambiguous candidate agreed on.
//
// The ICE default applies only when all four say nothing: no powertrain on the
// style row, none on the matched trim, none shared by the template, and a name
// carrying no hint. That is a genuinely unknown powertrain, and ICE remains
// the answer there for compatibility with what this endpoint has always
// served.
func powertrainForStyle(resolved services.Resolved, styleName string, styleMetadata []byte) string {
	if pt := gjson.GetBytes(styleMetadata, common.PowerTrainType).String(); pt != "" {
		return pt
	}
	resolvedPT, _ := resolved.Attributes[common.PowerTrainType].(string)
	if resolved.Quality == services.MatchExact && resolvedPT != "" {
		return resolvedPT
	}
	if pt := powertrainFromStyleName(styleName); pt != "" {
		return pt
	}
	if resolvedPT != "" {
		return resolvedPT
	}
	return models.PowertrainICE
}

// powertrainFromStyleName reads a powertrain out of a style name, or "" when
// the name says nothing. It is a heuristic over marketing names and is used
// only where nothing in the catalog answers: a trim that matched, or a
// template whose trims all agree, is always believed over it.
func powertrainFromStyleName(styleName string) string {
	name := strings.ToLower(styleName)
	switch {
	case strings.Contains(name, "phev"):
		return models.PowertrainPHEV
	case strings.Contains(name, "hev"):
		return models.PowertrainHEV
	case strings.Contains(name, "plug-in"):
		return models.PowertrainPHEV
	case strings.Contains(name, "hybrid"):
		return models.PowertrainHEV
	case strings.Contains(name, "electric"):
		return models.PowertrainBEV
	case strings.Contains(name, "4xe"):
		return models.PowertrainPHEV
	case strings.Contains(name, "energi"):
		return models.PowertrainPHEV
	}
	return ""
}

// convertTemplateAttributesToDeviceAttributes adapts a template's typed
// attribute map to the legacy DeviceTypeAttributeEditor shape this handler's
// response is built from.
func convertTemplateAttributesToDeviceAttributes(attributes map[string]any) []coremodels.DeviceTypeAttributeEditor {
	dta := make([]coremodels.DeviceTypeAttributeEditor, 0, len(attributes))
	for name, value := range attributes {
		dta = append(dta, coremodels.DeviceTypeAttributeEditor{
			Name:        name,
			Label:       name,
			Description: name,
			Value:       fmt.Sprint(value),
		})
	}
	return dta
}
