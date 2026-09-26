package queries

import (
	"context"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/logfields"
	vinutils "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type UpsertDecodingQueryHandler struct {
	dbs                            func() *db.ReaderWriter
	logger                         *zerolog.Logger
	deviceDefinitionCatalogService gateways.DeviceDefinitionCatalogService
}

type UpsertDecodingQuery struct {
	VIN          string `json:"vin"`
	DefinitionID string `json:"definitionId"`
}

func (*UpsertDecodingQuery) Key() string { return "UpsertDecodingQuery" }

func NewUpsertDecodingQueryHandler(dbs func() *db.ReaderWriter,
	logger *zerolog.Logger,
	deviceDefinitionCatalogService gateways.DeviceDefinitionCatalogService) UpsertDecodingQueryHandler {
	return UpsertDecodingQueryHandler{
		dbs:                            dbs,
		logger:                         logger,
		deviceDefinitionCatalogService: deviceDefinitionCatalogService,
	}
}

func (dc UpsertDecodingQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*UpsertDecodingQuery)
	if len(qry.VIN) < 10 || len(qry.VIN) > 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
	vin := vinutils.VIN(qry.VIN)
	wmi := vin.Wmi()

	localLog := dc.logger.With().
		Str("vin", vin.String()).
		Str("handler", query.Key()).
		Logger()

	// check if the definition id exists in the catalog. GetTemplateByID's
	// error is only ever "does not exist" when it is ErrTemplateNotFound
	// (checked by identity, never by message); anything else -- a catalog
	// outage, a timeout, a decode failure -- is an internal error and must
	// not be reported as a missing definition.
	dd, manuf, err := dc.deviceDefinitionCatalogService.GetTemplateByID(ctx, qry.DefinitionID)
	if err != nil {
		if errors.Is(err, gateways.ErrTemplateNotFound) {
			return nil, &exceptions.NotFoundError{Err: fmt.Errorf("device definition not found in catalog: %s: %w", qry.DefinitionID, err)}
		}
		return nil, &exceptions.InternalError{Err: fmt.Errorf("failed to find device definition by id %s when upserting vin decoding: %w", qry.DefinitionID, err)}
	}
	manufacturerName, err := manufacturerNameForTemplate(ctx, dc.deviceDefinitionCatalogService, dd, manuf)
	if err != nil {
		return nil, err
	}
	//upsert the vin
	vinNumber := &models.VinNumber{
		Vin:              qry.VIN,
		Wmi:              null.StringFrom(wmi),
		DecodeProvider:   null.StringFrom("manual entry"),
		Year:             dd.Year,
		DefinitionID:     dd.ID,
		ManufacturerName: manufacturerName,
	}
	if vin.IsValidVIN() {
		vinNumber.VDS = null.StringFrom(vin.VDS())
		vinNumber.CheckDigit = null.StringFrom(vin.CheckDigit())
		vinNumber.SerialNumber = vin.SerialNumber()
		vinNumber.Vis = null.StringFrom(vin.VIS())
	}

	err = vinNumber.Upsert(ctx, dc.dbs().Writer, true, []string{"vin"}, boil.Infer(), boil.Infer())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to upsert vin number %s for manual update", qry.VIN)
	}
	localLog.Info().Str(logfields.VIN, qry.VIN).Str(logfields.FunctionName, qry.Key()).Msg("manually upserted new vin number")
	return nil, nil
}

// manufacturerNameForTemplate names the manufacturer a template belongs to.
//
// A known token id keeps identity as the source of the name. A template's
// tokenId is optional, though, and GetTemplateByID reports an absent one as
// nil: looking that up as manufacturer 0 failed the upsert with an untyped
// error and never wrote the vin_numbers row. The template's own manufacturer
// name, which the schema requires, is used instead. A template with neither is
// reported as not found rather than written with an empty manufacturer.
func manufacturerNameForTemplate(ctx context.Context, catalog gateways.DeviceDefinitionCatalogService, tmpl *coremodels.Template, tokenID *big.Int) (string, error) {
	if tokenID == nil {
		if tmpl.Manufacturer.Name == "" {
			return "", &exceptions.NotFoundError{Err: fmt.Errorf("device definition %s has no manufacturer token id or name", tmpl.ID)}
		}
		return tmpl.Manufacturer.Name, nil
	}
	name, err := catalog.GetManufacturerNameByID(ctx, tokenID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to find manufacturer name by id %s when upserting vin decoding", tokenID)
	}
	return name, nil
}
