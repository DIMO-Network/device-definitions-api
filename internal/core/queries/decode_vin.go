package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/volatiletech/null/v8"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DecodeVINQueryHandler struct {
	dbs                func() *db.ReaderWriter
	vinDecodingService services.VINDecodingService
	logger             *zerolog.Logger
	repository         repositories.DeviceDefinitionRepository
	vinRepository      repositories.VINRepository
}

type DecodeVINQuery struct {
	VIN string `json:"vin"`
}

func (*DecodeVINQuery) Key() string { return "DecodeVINQuery" }

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter, vinDecodingService services.VINDecodingService,
	vinRepository repositories.VINRepository,
	repository repositories.DeviceDefinitionRepository, logger *zerolog.Logger) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		dbs:                dbs,
		vinDecodingService: vinDecodingService,
		logger:             logger,
		repository:         repository,
		vinRepository:      vinRepository,
	}
}

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*DecodeVINQuery)
	if len(qry.VIN) != 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
	resp := &p_grpc.DecodeVinResponse{}
	// get the year
	vin := shared.VIN(qry.VIN)
	year := int32(vin.Year()) // needs to be updated for newer years
	wmi := vin.Wmi()

	localLog := dc.logger.With().
		Str("vin", vin.String()).
		Str("handler", query.Key()).
		Str("year", fmt.Sprintf("%d", year)).
		Logger()

	vinDecodeNumber, err := models.FindVinNumber(ctx, dc.dbs().Reader, vin.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			localLog.Debug().Msg("No rows to decode vin numbers from db")
		}
	}

	if vinDecodeNumber != nil {

		resp.DeviceMakeId = vinDecodeNumber.DeviceMakeID
		resp.Year = year
		resp.DeviceDefinitionId = vinDecodeNumber.DeviceDefinitionID
		resp.DeviceStyleId = vinDecodeNumber.StyleID.String

		return resp, nil
	}

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		return nil, err
	}

	vinInfo, err := dc.vinDecodingService.GetVIN(vin.String(), dt)
	if err != nil {
		localLog.Err(err).Msg("failed to decode vin from drivly")
		return resp, nil
	}

	dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
	if err != nil {
		dc.logger.Error().Str("vin", vin.String()).Msgf("invalid vin %s", vin.String())
		return resp, nil
	}

	resp.Year = int32(vin.Year())
	resp.DeviceMakeId = dbWMI.DeviceMakeID

	// now match the model for the dd id
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbWMI.DeviceMakeID),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
		One(ctx, dc.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dd, err = dc.repository.GetOrCreate(ctx,
				vinInfo.Source,
				common.SlugString(vinInfo.Model+vinInfo.Year),
				dbWMI.DeviceMakeID,
				vinInfo.Model,
				int(year),
				common.DefaultDeviceType,
				vinInfo.MetaData,
				true,
				"")
			if err != nil {
				return nil, err
			}
			localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", common.SlugString(vinInfo.Model))
		} else {
			return nil, err
		}
	}
	if dd != nil {
		resp.DeviceDefinitionId = dd.ID
		// match style
		style, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
			models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName)).One(ctx, dc.dbs().Reader)
		if errors.Is(err, sql.ErrNoRows) {
			// insert
			style = &models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: dd.ID,
				Name:               vinInfo.StyleName,
				ExternalStyleID:    common.SlugString(vinInfo.StyleName),
				Source:             vinInfo.Source,
				SubModel:           vinInfo.SubModel,
			}
			_ = style.Insert(ctx, dc.dbs().Writer, boil.Infer())
			localLog.Info().Msgf("creating new device_style as did not find one for: %s", common.SlugString(vinInfo.StyleName))
		}
		resp.DeviceStyleId = style.ID
		// set the dd metadata if nothing there
		if !gjson.GetBytes(dd.Metadata.JSON, dt.Metadatakey).Exists() {
			// todo - future: merge metadata properties. Also set style specific metadata - multiple places
			dd.Metadata = vinInfo.MetaData
			_, _ = dd.Update(ctx, dc.dbs().Writer, boil.Whitelist(models.DeviceDefinitionColumns.Metadata, models.DeviceDefinitionColumns.UpdatedAt))
		}
		// todo- future: add powertrain - but this can be style specific
	}

	vinDecodeNumber = &models.VinNumber{
		Vin:                vin.String(),
		DeviceDefinitionID: dd.ID,
		DeviceMakeID:       dd.DeviceMakeID,
		StyleID:            null.StringFrom(resp.DeviceStyleId),
		Wmi:                wmi,
		VDS:                vin.VDS(),
		Vis:                vin.VIS(),
		CheckDigit:         vin.CheckDigit(),
		SerialNumber:       vin.SerialNumber(),
	}
	if err = vinDecodeNumber.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
		localLog.Err(err).
			Str("vin", vin.String()).
			Str("device_definition_id", dd.ID).
			Str("device_make_id", dd.DeviceMakeID).
			Msgf("failed to insert vin_numbers: %s", vin.String())
	}

	return resp, nil
}
