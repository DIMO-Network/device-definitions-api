package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/segmentio/ksuid"
	"strconv"

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
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DecodeVINQueryHandler struct {
	dbs                func() *db.ReaderWriter
	vinDecodingService services.VINDecodingService
	logger             *zerolog.Logger
	ddRepository       repositories.DeviceDefinitionRepository
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
		ddRepository:       repository,
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
	year := int32(vin.Year())
	wmi := vin.Wmi()

	localLog := dc.logger.With().
		Str("vin", vin.String()).
		Str("handler", query.Key()).
		Str("vin_year", fmt.Sprintf("%d", year)).
		Logger()

	// todo if year is 0, prefer vincario for decode, still send it through.
	if year == 0 {
		localLog.Warn().Msgf("could not decode vin. invalid vin encountered")
		return nil, fmt.Errorf("invalid vin encountered: %s", vin.String())
	}
	resp.Year = int32(vin.Year())

	vinDecodeNumber, err := models.FindVinNumber(ctx, dc.dbs().Reader, vin.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			localLog.Debug().Str("vin", vin.String()).Msg("no existing vin match found")
		}
	}

	if vinDecodeNumber != nil {
		resp.DeviceMakeId = vinDecodeNumber.DeviceMakeID
		//resp.Year = vinDecodeNumber.Year
		resp.DeviceDefinitionId = vinDecodeNumber.DeviceDefinitionID
		resp.DeviceStyleId = vinDecodeNumber.StyleID.String
		//resp.Source = vinDecodeNumber.DecodeProvider

		return resp, nil
	}

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		return nil, err
	}

	// future: see if we can self decode model based on data we have before calling external decode WMI and VDS. Only thing is we won't get the style.

	vinInfo, err := dc.vinDecodingService.GetVIN(vin.String(), dt)
	if err != nil {
		localLog.Err(err).Msgf("failed to decode vin from %s", vinInfo.Source)
		return resp, err
	}
	localLog = localLog.With().Str("decode_source", vinInfo.Source).Logger()

	if len(vinInfo.Model) == 0 {
		localLog.Warn().Msg("decoded model name must have a minimum of 1 characters.")
		return nil, errors.New("decoded model name is blank")
	}

	dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
	if err != nil {
		dc.logger.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
		return resp, nil
	}
	resp.DeviceMakeId = dbWMI.DeviceMakeID
	resp.Source = vinInfo.Source
	if atoi, err := strconv.Atoi(vinInfo.Year); err == nil {
		year = int32(atoi)
		resp.Year = int32(atoi)
	}

	// now match the model for the dd id
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbWMI.DeviceMakeID),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
		One(ctx, dc.dbs().Reader)
	if err != nil {
		// create DD if does not exist
		if errors.Is(err, sql.ErrNoRows) {
			dd, err = dc.ddRepository.GetOrCreate(ctx,
				vinInfo.Source,
				common.SlugString(vinInfo.Model+vinInfo.Year),
				dbWMI.DeviceMakeID,
				vinInfo.Model,
				int(year),
				common.DefaultDeviceType,
				vinInfo.MetaData,
				true,
				nil)
			if err != nil {
				return nil, err
			}
			localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", common.SlugString(vinInfo.Model))
		} else {
			return nil, err
		}
	}
	if dd == nil {
		return nil, errors.New("could not get or create device_definition")
	}
	resp.DeviceDefinitionId = dd.ID
	// match style - only process style if name is longer than 1
	if len(vinInfo.StyleName) < 2 {
		localLog.Warn().Msgf("decoded style name too short: %s must have a minimum of 2 characters.", vinInfo.StyleName)
	} else {
		style, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
			models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName)).One(ctx, dc.dbs().Reader)
		if errors.Is(err, sql.ErrNoRows) {
			// insert, if fails doesn't matter - continue just don't return the style_id
			style = &models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: dd.ID,
				Name:               vinInfo.StyleName,
				ExternalStyleID:    common.SlugString(vinInfo.StyleName),
				Source:             vinInfo.Source,
				SubModel:           vinInfo.SubModel,
			}
			err := style.Insert(ctx, dc.dbs().Writer, boil.Infer())
			if err == nil {
				localLog.Info().Msgf("creating new device_style as did not find one for: %s", common.SlugString(vinInfo.StyleName))
				resp.DeviceStyleId = style.ID
			}
		} else if err == nil {
			resp.DeviceStyleId = style.ID
		}
	}

	// set the dd metadata if nothing there, if fails just continue
	if !gjson.GetBytes(dd.Metadata.JSON, dt.Metadatakey).Exists() {
		// todo - future: merge metadata properties. Also set style specific metadata - multiple places
		dd.Metadata = vinInfo.MetaData
		_, _ = dd.Update(ctx, dc.dbs().Writer, boil.Whitelist(models.DeviceDefinitionColumns.Metadata, models.DeviceDefinitionColumns.UpdatedAt))
		// todo- future: add powertrain - but this can be style specific - vincario gets us primary FuelType
	}
	// insert vin_numbers
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
		//DecodeProvider:     vinInfo.Source,
	}
	if err = vinDecodeNumber.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
		localLog.Err(err).
			Str("device_definition_id", dd.ID).
			Str("device_make_id", dd.DeviceMakeID).
			Msg("failed to insert to vin_numbers")
	}

	return resp, nil
}
