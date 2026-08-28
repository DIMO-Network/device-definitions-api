//nolint:tagliatelle
package queries

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
)

type GetDeviceDefinitionByDynamicFilterQuery struct {
	DefinitionID    string   `json:"definition_id"`
	Year            int      `json:"year"`
	Model           string   `json:"model"`
	VerifiedVinList []string `json:"verified_vin_list"`
	PageIndex       int      `json:"page_index"`
	PageSize        int      `json:"page_size"`
	MakeSlug        string
}

type DeviceDefinitionQueryResponse struct {
	ID           string      `json:"id"`
	NameSlug     string      `json:"name_slug"`
	Model        string      `json:"model"`
	Year         int         `json:"year"`
	ImageURL     null.String `json:"image_url,omitempty"`
	CreatedAt    time.Time   `json:"created_at,omitempty"`
	UpdatedAt    time.Time   `json:"updated_at,omitempty"`
	Metadata     null.JSON   `json:"metadata"`
	Source       null.String `json:"source"`
	Verified     bool        `json:"verified"`
	ExternalID   null.String `json:"external_id"`
	DeviceMakeID string      `json:"device_make_id"`
	Make         string      `json:"make"`
	ExternalIDs  null.JSON   `json:"external_ids"`
}

func (*GetDeviceDefinitionByDynamicFilterQuery) Key() string {
	return "GetDeviceDefinitionByDynamicFilterQuery"
}

type GetDeviceDefinitionByDynamicFilterQueryHandler struct {
	DBS        func() *db.ReaderWriter
	catalogSvc gateways.DeviceDefinitionCatalogService
}

func NewGetDeviceDefinitionByDynamicFilterQueryHandler(dbs func() *db.ReaderWriter, catalogSvc gateways.DeviceDefinitionCatalogService) GetDeviceDefinitionByDynamicFilterQueryHandler {
	return GetDeviceDefinitionByDynamicFilterQueryHandler{
		DBS:        dbs,
		catalogSvc: catalogSvc,
	}
}

func (ch GetDeviceDefinitionByDynamicFilterQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByDynamicFilterQuery)

	if len(qry.DefinitionID) > 1 {
		dd, _, err := ch.catalogSvc.GetTemplateByID(ctx, qry.DefinitionID)
		if err != nil {
			return nil, err
		}
		dds := make([]DeviceDefinitionQueryResponse, 1)
		dds[0] = ch.buildDeviceDefinitionQueryResponseFromTemplate(dd)
		return dds, nil
	}

	manufacturerID := types.NullDecimal{}

	if len(qry.MakeSlug) > 1 {
		manufacturer, err := ch.catalogSvc.GetManufacturer(qry.MakeSlug)
		if err != nil {
			return nil, err
		}
		manufacturerID = types.NewNullDecimal(decimal.New(int64(manufacturer.TokenID), 0))
	}

	definitions, err := ch.catalogSvc.GetDeviceDefinitions(ctx, manufacturerID, "", qry.Model, qry.Year, int32(qry.PageIndex), int32(qry.PageSize))
	if err != nil {
		return nil, err
	}

	dd := make([]DeviceDefinitionQueryResponse, len(definitions))
	for i, item := range definitions {
		dd[i] = ch.buildDeviceDefinitionQueryResponse(&item)
	}

	return dd, err

}

func (ch GetDeviceDefinitionByDynamicFilterQueryHandler) buildDeviceDefinitionQueryResponse(dd *models.DeviceDefinitionTablelandModel) DeviceDefinitionQueryResponse {
	if dd == nil {
		return DeviceDefinitionQueryResponse{}
	}
	split := strings.Split(dd.ID, "_")
	manufacturerSlug := split[0]
	manufacturer, _ := ch.catalogSvc.GetManufacturer(manufacturerSlug)
	mdStr := []byte("{}")
	if dd.Metadata != nil {
		mdStr, _ = json.Marshal(dd.Metadata)
	}

	return DeviceDefinitionQueryResponse{
		ID:           dd.ID,
		NameSlug:     dd.ID,
		Model:        dd.Model,
		Year:         dd.Year,
		ImageURL:     null.StringFrom(dd.ImageURI),
		Verified:     true,
		DeviceMakeID: strconv.Itoa(manufacturer.TokenID),
		Make:         manufacturer.Name,
		Metadata:     null.JSONFrom(mdStr),
	}
}

// buildDeviceDefinitionQueryResponseFromTemplate is
// buildDeviceDefinitionQueryResponse for the catalog's template-shaped read
// path: same response, sourced from a *models.Template rather than the flat
// tableland model GetDeviceDefinitions still returns.
func (ch GetDeviceDefinitionByDynamicFilterQueryHandler) buildDeviceDefinitionQueryResponseFromTemplate(tmpl *models.Template) DeviceDefinitionQueryResponse {
	if tmpl == nil {
		return DeviceDefinitionQueryResponse{}
	}
	split := strings.Split(tmpl.ID, "_")
	manufacturerSlug := split[0]
	manufacturer, _ := ch.catalogSvc.GetManufacturer(manufacturerSlug)
	mdStr := []byte("{}")
	if tmpl.Attributes != nil {
		mdStr, _ = json.Marshal(tmpl.Attributes)
	}

	return DeviceDefinitionQueryResponse{
		ID:           tmpl.ID,
		NameSlug:     tmpl.ID,
		Model:        tmpl.Model,
		Year:         tmpl.Year,
		ImageURL:     null.StringFrom(tmpl.ImageURI),
		Verified:     true,
		DeviceMakeID: strconv.Itoa(manufacturer.TokenID),
		Make:         manufacturer.Name,
		Metadata:     null.JSONFrom(mdStr),
	}
}
