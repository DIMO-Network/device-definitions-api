package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
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

	// check if the definition id exists on chain
	dd, manuf, err := dc.deviceDefinitionCatalogService.GetTemplateByID(ctx, qry.DefinitionID)
	if err == nil && dd == nil {
		return nil, &exceptions.NotFoundError{Err: fmt.Errorf("device definition not found in catalog: %s", qry.DefinitionID)}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find device definition by id %s when upserting vin decoding", qry.DefinitionID)
	}
	manufacturerName, err := dc.deviceDefinitionCatalogService.GetManufacturerNameByID(ctx, manuf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find manufacturer name by id %s when upserting vin decoding", manuf)
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
