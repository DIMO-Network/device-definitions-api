//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/sjson"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DecodeVINQueryHandler struct {
	dbs                            func() *db.ReaderWriter
	vinDecodingService             services.VINDecodingService
	logger                         *zerolog.Logger
	ddRepository                   repositories.DeviceDefinitionRepository
	vinRepository                  repositories.VINRepository
	fuelAPIService                 gateways.FuelAPIService
	powerTrainTypeService          services.PowerTrainTypeService
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService
}

type DecodeVINQuery struct {
	VIN                string `json:"vin"`
	KnownModel         string `json:"knownModel"`
	KnownYear          int32  `json:"knownYear"`
	Country            string `json:"country"`
	DeviceDefinitionID string `json:"device_definition_id"`
}

func (*DecodeVINQuery) Key() string { return "DecodeVINQuery" }

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter, vinDecodingService services.VINDecodingService,
	vinRepository repositories.VINRepository,
	repository repositories.DeviceDefinitionRepository, logger *zerolog.Logger,
	fuelAPIService gateways.FuelAPIService,
	powerTrainTypeService services.PowerTrainTypeService,
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		dbs:                            dbs,
		vinDecodingService:             vinDecodingService,
		logger:                         logger,
		ddRepository:                   repository,
		vinRepository:                  vinRepository,
		fuelAPIService:                 fuelAPIService,
		powerTrainTypeService:          powerTrainTypeService,
		deviceDefinitionOnChainService: deviceDefinitionOnChainService,
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
		Str("vinYear", fmt.Sprintf("%d", resp.Year)).
		Str("knownModel", qry.KnownModel).
		Str("knownYear", strconv.Itoa(int(qry.KnownYear))).
		Str("country", qry.Country).
		Logger()

	const (
		VinRequests              = "VIN_All_Request"
		VinSuccess               = "VIN_Success_Request"
		VinExists                = "VIN_Exists_Request"
		VinErrors                = "VIN_Error_Request"
		DeviceDefinitionOverride = "Device_Definition_Override"
	)

	metrics.Success.With(prometheus.Labels{"method": VinRequests}).Inc()
	txVinNumbers, err := dc.dbs().Writer.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, errors.Wrap(err, "error when beginning transaction")
	}
	defer txVinNumbers.Rollback() //nolint
	vinDecodeNumber, err := models.VinNumbers(
		models.VinNumberWhere.Vin.EQ(vin.String()),
		qm.Load(models.VinNumberRels.DeviceDefinition)).
		One(ctx, txVinNumbers)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errors.Wrap(err, "error when querying for existing VIN number")
	}
	if vinDecodeNumber != nil {
		resp.DeviceMakeId = vinDecodeNumber.DeviceMakeID
		resp.Year = int32(vinDecodeNumber.Year)
		resp.DeviceDefinitionId = vinDecodeNumber.DeviceDefinitionID
		resp.DeviceStyleId = vinDecodeNumber.StyleID.String
		resp.Source = vinDecodeNumber.DecodeProvider.String
		resp.NameSlug = vinDecodeNumber.R.DeviceDefinition.NameSlug

		pt, err := dc.powerTrainTypeService.ResolvePowerTrainType(ctx, "", "", &vinDecodeNumber.DeviceDefinitionID, vinDecodeNumber.DrivlyData, vinDecodeNumber.VincarioData)
		if err != nil {
			pt = coremodels.ICE.String()
		}
		resp.Powertrain = pt

		metrics.Success.With(prometheus.Labels{"method": VinExists}).Inc()

		return resp, nil
	}

	// If DeviceDefinitionID passed in, override VIN decoding
	localLog.Info().Msgf("Start Decode VIN for vin %s and device definition %s", vin.String(), qry.DeviceDefinitionID)
	if len(qry.DeviceDefinitionID) > 0 {
		dd, err := dc.ddRepository.GetByID(ctx, qry.DeviceDefinitionID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get device definition id")
		}

		dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, dd.R.DeviceMake.Name)
		if err != nil {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			localLog.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
			return resp, nil
		}

		// insert vin_numbers
		vinDecodeNumber = &models.VinNumber{
			Vin:                vin.String(),
			DeviceDefinitionID: dd.ID,
			DeviceMakeID:       dd.DeviceMakeID,
			Wmi:                dbWMI.Wmi,
			VDS:                vin.VDS(),
			Vis:                vin.VIS(),
			CheckDigit:         vin.CheckDigit(),
			SerialNumber:       vin.SerialNumber(),
			DecodeProvider:     null.StringFrom("manual"),
			Year:               int(dd.Year),
		}

		// no style, maybe for future way to pick the Style from Admin

		// note we use a transaction here all throughout and commit at the end
		if err = vinDecodeNumber.Insert(ctx, txVinNumbers, boil.Infer()); err != nil {
			localLog.Err(err).
				Str("device_definition_id", dd.ID).
				Str("device_make_id", dd.DeviceMakeID).
				Msg("failed to insert to vin_numbers")
		}
		err = txVinNumbers.Commit()
		if err != nil {
			return nil, errors.Wrap(err, "error when commiting transaction for inserting vin_number")
		}

		resp.DeviceMakeId = vinDecodeNumber.DeviceMakeID
		resp.Year = int32(vinDecodeNumber.Year)
		resp.DeviceDefinitionId = vinDecodeNumber.DeviceDefinitionID
		resp.Source = vinDecodeNumber.DecodeProvider.String
		pt, err := dc.powerTrainTypeService.ResolvePowerTrainType(ctx, "", "", &vinDecodeNumber.DeviceDefinitionID, null.JSON{}, null.JSON{})
		if err != nil {
			pt = coremodels.ICE.String()
		}
		resp.Powertrain = pt
		resp.NameSlug = dd.NameSlug

		metrics.Success.With(prometheus.Labels{"method": DeviceDefinitionOverride}).Inc()

		return resp, nil
	}

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errors.Wrap(err, "failed to get device_type")
	}
	// future: see if we can self decode model based on data we have before calling external decode WMI and VDS. Only thing is we won't get the style.

	var vinInfo = &coremodels.VINDecodingInfoData{}
	// if year is 0 or way in future, prefer datgroup, autoiso and vincario for decode, since most likely non USA.
	if resp.Year == 0 || resp.Year > int32(time.Now().Year()+1) {
		localLog.Info().Msgf("encountered vin with non-standard year digit")
		vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.DATGroupProvider, qry.Country)
		if err != nil {
			localLog.Err(err).Msg("failed to GetVIN with DATGroupProvider")
			vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.VincarioProvider, qry.Country)
		}
	} else {
		vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.AllProviders, qry.Country) // this will try drivly first
	}

	// if no luck decoding VIN, try buildingVinInfo from known data passed in
	if err != nil {
		if len(qry.KnownModel) > 0 && qry.KnownYear > 0 {
			// note if this is successful, err gets set to nil
			vinInfo, err = dc.vinInfoFromKnown(vin, qry.KnownModel, qry.KnownYear)
		}
	}

	if vinInfo != nil {
		localLog = localLog.With().Str("decode_source", string(vinInfo.Source)).Logger()
	}

	if err != nil || vinInfo == nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		if err == nil {
			err = errors.New("failed to decode, vinInfo is nil")
		}
		localLog.Err(err).Msgf("failed to decode vin from provider, country: %s", qry.Country)
		return nil, err
	}

	dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		dc.logger.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
		return resp, nil
	}
	resp.DeviceMakeId = dbWMI.DeviceMakeID
	resp.Source = string(vinInfo.Source)
	resp.Year = vinInfo.Year

	// now match the model for the dd id
	// todo, next iteration: have this query tableland. underlying method will need to get the table id, we may already have this in part in identity-api
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbWMI.DeviceMakeID),
		models.DeviceDefinitionWhere.Year.EQ(int16(resp.Year)),
		models.DeviceDefinitionWhere.ModelSlug.EQ(shared.SlugString(vinInfo.Model))).
		One(ctx, dc.dbs().Reader)

	ddExists := true
	if err != nil {
		// create DD if does not exist, metadata will only be set on create
		if errors.Is(err, sql.ErrNoRows) {
			// towards the end we create the record on-chain, this should be removed eventually
			dd, err = dc.ddRepository.GetOrCreate(ctx, txVinNumbers,
				string(vinInfo.Source),
				shared.SlugString(vinInfo.Model+strconv.Itoa(int(vinInfo.Year))),
				dbWMI.DeviceMakeID,
				vinInfo.Model,
				int(resp.Year),
				common.DefaultDeviceType,
				vinInfo.MetaData,
				true,
				nil)
			if err != nil {
				metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
				return nil, errors.Wrap(err, "error creating new device definition from decoded vin")
			}
			ddExists = false
			localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", shared.SlugString(vinInfo.Model))
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
	resp.NameSlug = dd.NameSlug

	// match style - only process style if name is longer than 1
	if len(vinInfo.StyleName) < 2 {
		localLog.Warn().Msgf("decoded style name too short: %s must have a minimum of 2 characters.", vinInfo.StyleName)
	} else {
		externalStyleID := shared.SlugString(vinInfo.StyleName)
		// see if match existing style exists. First search is based on db device_definition_style_idx
		style, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
			models.DeviceStyleWhere.Source.EQ(string(vinInfo.Source)),
			models.DeviceStyleWhere.ExternalStyleID.EQ(externalStyleID)).One(ctx, dc.dbs().Reader)
		if errors.Is(err, sql.ErrNoRows) {
			style, err = models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
				models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName)).One(ctx, dc.dbs().Reader)
		}

		if errors.Is(err, sql.ErrNoRows) {
			style = &models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: dd.ID,
				Name:               vinInfo.StyleName,
				ExternalStyleID:    externalStyleID,
				Source:             string(vinInfo.Source),
				SubModel:           vinInfo.SubModel,
				Metadata:           vinInfo.MetaData,
			}
			// style level powertrain
			pt := dc.powerTrainTypeService.ResolvePowerTrainFromVinInfo(vinInfo)
			if pt != "" {
				metadataWithPT, metadataErr := sjson.SetBytes(vinInfo.MetaData.JSON, common.PowerTrainType, pt)
				if metadataErr == nil {
					style.Metadata = null.JSONFrom(metadataWithPT)
				}
				// todo unit test for this getting set in this scenario
				resp.Powertrain = pt
			}
			errStyle := style.Insert(ctx, txVinNumbers, boil.Infer())
			if errStyle != nil {
				localLog.Err(errStyle).Msgf("error creating style with values: %+v", style)
				return nil, errStyle
			}
			localLog.Info().Msgf("creating new device_style as did not find one for: %s", shared.SlugString(vinInfo.StyleName))
			resp.DeviceStyleId = style.ID

		} else if err == nil {
			resp.DeviceStyleId = style.ID
		}
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
	if vinInfo.Source == coremodels.AutoIsoProvider && len(vinInfo.Raw) > 0 {
		vinDecodeNumber.AutoisoData = null.JSONFrom(vinInfo.Raw)
	}
	if vinInfo.Source == coremodels.DATGroupProvider && len(vinInfo.Raw) > 0 {
		vinDecodeNumber.DatgroupData = null.JSONFrom(vinInfo.Raw)
	}

	localLog.Info().Str("device_definition_id", dd.ID).
		Str("device_make_id", dd.DeviceMakeID).
		Str("style_id", resp.DeviceStyleId).
		Str("wmi", wmi).
		Str("vds", vin.VDS()).
		Str("vis", vin.VIS()).
		Str("check_digit", vin.CheckDigit()).Msgf("decoded vin ok with: %s", vinInfo.Source)
	// note we use a transaction here all throughout and commit at the end
	if err = vinDecodeNumber.Insert(ctx, txVinNumbers, boil.Infer()); err != nil {
		localLog.Err(err).
			Str("device_definition_id", dd.ID).
			Str("device_make_id", dd.DeviceMakeID).
			Msg("failed to insert to vin_numbers")
	}
	err = txVinNumbers.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "error when commiting transaction for inserting vin_number")
	}

	// resolve images
	images, _ := models.Images(models.ImageWhere.DeviceDefinitionID.EQ(dd.ID)).All(ctx, dc.dbs().Reader)
	localLog.Debug().Msgf("Current Images : %d", len(images))

	if len(images) == 0 {
		err = dc.associateImagesToDeviceDefinition(ctx, dd.ID, vinInfo.Make, vinInfo.Model, int(resp.Year), 2, 2)
		if err != nil {
			localLog.Err(err).Send()
		}

		err = dc.associateImagesToDeviceDefinition(ctx, dd.ID, vinInfo.Make, vinInfo.Model, int(resp.Year), 2, 6)
		if err != nil {
			localLog.Err(err).Send()
		}
	}

	if !ddExists {
		dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.EQ(dd.ID),
			qm.Load(models.DeviceDefinitionRels.DeviceStyles),
			qm.Load(models.DeviceDefinitionRels.DeviceType),
			qm.Load(models.DeviceDefinitionRels.DeviceMake),
			qm.Load(models.DeviceDefinitionRels.Images)).One(ctx, dc.dbs().Reader)
		if err != nil {
			return nil, errors.Wrap(err, "error when get dd for update powertraintype")
		}

		metadataKey := dd.R.DeviceType.Metadatakey
		var metadataAttributes map[string]any

		if err := dd.Metadata.Unmarshal(&metadataAttributes); err == nil {
			metaData := make(map[string]interface{})
			if metadataAttributes == nil {
				metadataAttributes = make(map[string]interface{})
				var deviceTypeAttributes map[string][]coremodels.GetDeviceTypeAttributeQueryResult
				if err := dd.R.DeviceType.Properties.Unmarshal(&deviceTypeAttributes); err == nil {
					for _, deviceAttribute := range deviceTypeAttributes["properties"] {
						metaData[deviceAttribute.Name] = deviceAttribute.DefaultValue
					}
				}

				metadataAttributes[metadataKey] = metaData
			}
		}

		if metadataAttributes != nil {
			if metadataValue, ok := metadataAttributes[metadataKey]; ok {
				for key, value := range metadataValue.(map[string]interface{}) {
					if key == common.PowerTrainType {
						powerTrainTypeValue := value
						if powerTrainTypeValue == nil || powerTrainTypeValue == "" {
							drivlyData := null.JSON{}
							vincarioData := null.JSON{}
							if vinInfo.Source == coremodels.DrivlyProvider {
								drivlyData = vinInfo.MetaData
							} else {
								vincarioData = vinInfo.MetaData
							}
							powerTrainTypeValue, err = dc.powerTrainTypeService.ResolvePowerTrainType(ctx, dd.R.DeviceMake.NameSlug, dd.ModelSlug, nil, drivlyData, vincarioData)
							if err != nil {
								dc.logger.Error().Err(err).Msg("Error when resolve Powertrain")
							}
						}

						metadataAttributes[metadataKey].(map[string]interface{})[common.PowerTrainType] = powerTrainTypeValue
					}
				}
			}
		}

		err = dd.Metadata.Marshal(metadataAttributes)
		if err != nil {
			return nil, err
		}

		dd.HardwareTemplateID = null.StringFrom(common.DefautlAutoPiTemplate)

		// Create DD onchain
		trx, err := dc.deviceDefinitionOnChainService.Create(ctx, *dd.R.DeviceMake, *dd)
		if err != nil {
			localLog.Err(err).Msg("failed to create or update DD on chain")
		}
		if err == nil {
			trxArray := strings.Split(*trx, ",")
			if dd.TRXHashHex != nil {
				dd.TRXHashHex = append(dd.TRXHashHex, trxArray...)
			} else {
				dd.TRXHashHex = trxArray
			}
		}

		if err = dd.Upsert(ctx, dc.dbs().Writer, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
			return nil, err
		}
		if trx != nil {
			resp.NewTrxHash = *trx
		}
	}

	metrics.Success.With(prometheus.Labels{"method": VinSuccess}).Inc()

	// if powertrain not set yet, try resolving for it
	if resp.Powertrain == "" {
		pt, _ := dc.powerTrainTypeService.ResolvePowerTrainType(ctx, "", "", &resp.DeviceDefinitionId, vinDecodeNumber.DrivlyData, vinDecodeNumber.VincarioData)
		resp.Powertrain = pt
	}

	return resp, nil
}

// vinInfoFromKnown builds a vininfo object based on one passed in with Make from vin WMI, and passed in model and year set
func (dc DecodeVINQueryHandler) vinInfoFromKnown(vin shared.VIN, knownModel string, knownYear int32) (*coremodels.VINDecodingInfoData, error) {
	vinInfo := &coremodels.VINDecodingInfoData{}
	vinInfo.VIN = vin.String()
	wmi, err := models.Wmis(models.WmiWhere.Wmi.EQ(vin.Wmi()),
		qm.Load(models.WmiRels.DeviceMake)).One(context.Background(), dc.dbs().Reader)
	if err != nil {
		return nil, errors.Wrap(err, "vinInfoFromKnown: could not get WMI from vin wmi "+vin.Wmi())
	}
	vinInfo.Make = wmi.R.DeviceMake.Name
	vinInfo.Year = knownYear
	vinInfo.Model = knownModel
	vinInfo.Source = "probably smartcar"

	if len(vinInfo.Model) == 0 || len(vinInfo.Make) == 0 || vinInfo.Year == 0 {
		return nil, fmt.Errorf("unable to decode from known info")
	}

	return vinInfo, nil
}

func (dc DecodeVINQueryHandler) associateImagesToDeviceDefinition(ctx context.Context, deviceDefinitionID, make, model string, year int, prodID int, prodFormat int) error {

	img, err := dc.fuelAPIService.FetchDeviceImages(make, model, year, prodID, prodFormat)
	if err != nil {
		dc.logger.Warn().Err(err).Msgf("unable to fetch device image for: %d %s %s", year, make, model)
		return nil
	}

	var p models.Image

	// loop through all img (color variations)
	for _, device := range img.Images {
		p.ID = ksuid.New().String()
		p.DeviceDefinitionID = deviceDefinitionID
		p.FuelAPIID = null.StringFrom(img.FuelAPIID)
		p.Width = null.IntFrom(img.Width)
		p.Height = null.IntFrom(img.Height)
		p.SourceURL = device.SourceURL
		//p.DimoS3URL = null.StringFrom("") // dont set it so it is null
		p.Color = device.Color
		p.NotExactImage = img.NotExactImage

		err = p.Upsert(ctx, dc.dbs().Writer, true, []string{models.ImageColumns.DeviceDefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
		if err != nil {
			dc.logger.Warn().Msgf("fail insert device image for: %s %d %s %s", deviceDefinitionID, year, make, model)
			continue
		}
	}

	return nil
}
