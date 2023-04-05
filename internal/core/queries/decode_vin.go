package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/prometheus/client_golang/prometheus"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/segmentio/ksuid"

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
	vin := shared.VIN(qry.VIN)
	resp.Year = int32(vin.Year())
	wmi := vin.Wmi()

	localLog := dc.logger.With().
		Str("vin", vin.String()).
		Str("handler", query.Key()).
		Str("vin_year", fmt.Sprintf("%d", resp.Year)).
		Logger()

	const (
		VinRequests = "VIN_All_Request"
		VinSuccess  = "VIN_Success_Request"
		VinErrors   = "VIN_Error_Request"
	)

	metrics.Success.With(prometheus.Labels{"method": VinRequests}).Inc()

	vinDecodeNumber, err := models.FindVinNumber(ctx, dc.dbs().Reader, vin.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, err
	}
	if vinDecodeNumber != nil {
		resp.DeviceMakeId = vinDecodeNumber.DeviceMakeID
		resp.Year = int32(vinDecodeNumber.Year)
		resp.DeviceDefinitionId = vinDecodeNumber.DeviceDefinitionID
		resp.DeviceStyleId = vinDecodeNumber.StyleID.String
		resp.Source = vinDecodeNumber.DecodeProvider.String

		metrics.Success.With(prometheus.Labels{"method": VinSuccess}).Inc()

		return resp, nil
	}

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, err
	}
	// future: see if we can self decode model based on data we have before calling external decode WMI and VDS. Only thing is we won't get the style.

	var vinInfo = &coremodels.VINDecodingInfoData{}
	// if year is 0, prefer vincario for decode, since most likely non USA.
	if resp.Year == 0 {
		localLog.Info().Msgf("encountered vin with non-standard year digit")
		vinInfo, err = dc.vinDecodingService.GetVIN(vin.String(), dt, coremodels.VincarioProvider)
	} else {
		vinInfo, err = dc.vinDecodingService.GetVIN(vin.String(), dt, coremodels.AllProviders)
	}

	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		localLog.Err(err).Msgf("failed to decode vin from provider %s", vinInfo.Source)
		return resp, err
	}
	localLog = localLog.With().Str("decode_source", string(vinInfo.Source)).Logger()

	if len(vinInfo.Model) == 0 {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		localLog.Warn().Msg("decoded model name must have a minimum of 1 characters.")
		return nil, errors.New("decoded model name is blank")
	}

	dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		dc.logger.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
		return resp, nil
	}
	resp.DeviceMakeId = dbWMI.DeviceMakeID
	resp.Source = string(vinInfo.Source)
	if atoi, err := strconv.Atoi(vinInfo.Year); err == nil {
		resp.Year = int32(atoi)
	}

	// now match the model for the dd id
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbWMI.DeviceMakeID),
		models.DeviceDefinitionWhere.Year.EQ(int16(resp.Year)),
		models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
		One(ctx, dc.dbs().Reader)
	if err != nil {
		// create DD if does not exist
		if errors.Is(err, sql.ErrNoRows) {
			dd, err = dc.ddRepository.GetOrCreate(ctx,
				string(vinInfo.Source),
				common.SlugString(vinInfo.Model+vinInfo.Year),
				dbWMI.DeviceMakeID,
				vinInfo.Model,
				int(resp.Year),
				common.DefaultDeviceType,
				vinInfo.MetaData,
				true,
				nil)
			if err != nil {
				metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
				return nil, err
			}
			localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", common.SlugString(vinInfo.Model))
		} else {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			return nil, err
		}
	}
	if dd == nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
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
				Source:             string(vinInfo.Source),
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
		Wmi:                wmi,
		VDS:                vin.VDS(),
		Vis:                vin.VIS(),
		CheckDigit:         vin.CheckDigit(),
		SerialNumber:       vin.SerialNumber(),
		DecodeProvider:     null.StringFrom(string(vinInfo.Source)),
		Year:               int(resp.Year),
	}
	if len(resp.DeviceStyleId) > 0 {
		vinDecodeNumber.StyleID = null.StringFrom(resp.DeviceStyleId)
	}
	if vinInfo.Source == coremodels.DrivlyProvider && len(vinInfo.Raw) > 0 {
		vinDecodeNumber.DrivlyData = null.JSONFrom(vinInfo.Raw)
	}
	if vinInfo.Source == coremodels.VincarioProvider && len(vinInfo.Raw) > 0 {
		vinDecodeNumber.VincarioData = null.JSONFrom(vinInfo.Raw)
	}

	localLog.Info().Str("device_definition_id", dd.ID).
		Str("device_make_id", dd.DeviceMakeID).
		Str("style_id", resp.DeviceStyleId).
		Str("wmi", wmi).
		Str("vds", vin.VDS()).
		Str("vis", vin.VIS()).
		Str("check_digit", vin.CheckDigit()).Msg("decoded vin ok")

	if err = vinDecodeNumber.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
		localLog.Err(err).
			Str("device_definition_id", dd.ID).
			Str("device_make_id", dd.DeviceMakeID).
			Msg("failed to insert to vin_numbers")
	}

	metrics.Success.With(prometheus.Labels{"method": VinSuccess}).Inc()

	return resp, nil
}
