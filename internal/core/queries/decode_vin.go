//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

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
	VIN          string `json:"vin"`
	KnownModel   string `json:"knownModel"`
	KnownYear    int32  `json:"knownYear"`
	Country      string `json:"country"`
	DefinitionID string `json:"definition_id"`
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
		models.VinNumberWhere.Vin.EQ(vin.String())).
		One(ctx, txVinNumbers)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errors.Wrap(err, "error when querying for existing VIN number")
	}
	// todo refactor: func hydrateResponseFromVinNumber(...)
	if vinDecodeNumber != nil {
		// get from tableland, probably don't need this here anymore
		//tblDef, errTbl := dc.deviceDefinitionOnChainService.GetDefinitionByID(ctx, vinDecodeNumber.DefinitionID, dc.dbs().Reader)
		// should mark this deprecated
		//resp.DeviceMakeId = vinDecodeNumber.DeviceMakeID
		resp.Manufacturer = vinDecodeNumber.ManufacturerName
		resp.Year = int32(vinDecodeNumber.Year)
		resp.DeviceStyleId = vinDecodeNumber.StyleID.String
		resp.Source = vinDecodeNumber.DecodeProvider.String
		resp.DefinitionId = vinDecodeNumber.DefinitionID
		split := strings.Split(vinDecodeNumber.DefinitionID, "_")
		if len(split) != 3 {
			return nil, errors.New("invalid definition ID encountered: " + vinDecodeNumber.DefinitionID)
		}
		pt, err := dc.powerTrainTypeService.ResolvePowerTrainType(split[0], split[1], vinDecodeNumber.DrivlyData, vinDecodeNumber.VincarioData)
		if err != nil {
			pt = coremodels.ICE.String()
		}
		resp.Powertrain = pt

		metrics.Success.With(prometheus.Labels{"method": VinExists}).Inc()

		return resp, nil
	}

	// If DeviceDefinitionID passed in, override VIN decoding
	localLog.Info().Msgf("Start Decode VIN for vin %s and device definition %s", vin.String(), qry.DefinitionID)
	if len(qry.DefinitionID) > 0 {
		tblDef, _, err := dc.deviceDefinitionOnChainService.GetDefinitionByID(ctx, qry.DefinitionID, dc.dbs().Reader)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get device definition id: %s", qry.DefinitionID)
		}
		makeSlug := strings.Split(tblDef.ID, "_")[0]
		dm, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(makeSlug)).One(ctx, dc.dbs().Reader)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get device make for: %s", qry.DefinitionID)
		}
		dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, dm.Name)
		if err != nil {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			localLog.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
			return resp, nil
		}

		// insert vin_numbers
		vinDecodeNumber = &models.VinNumber{
			Vin:              vin.String(),
			ManufacturerName: dm.Name,
			Wmi:              dbWMI.Wmi,
			VDS:              vin.VDS(),
			Vis:              vin.VIS(),
			CheckDigit:       vin.CheckDigit(),
			SerialNumber:     vin.SerialNumber(),
			DecodeProvider:   null.StringFrom("manual"),
			Year:             tblDef.Year,
			DefinitionID:     tblDef.ID,
		}

		split := strings.Split(vinDecodeNumber.DefinitionID, "_")
		if len(split) != 3 {
			return nil, errors.New("invalid definition ID encountered: " + vinDecodeNumber.DefinitionID)
		}

		// no style, maybe for future way to pick the Style from Admin

		// note we use a transaction here all throughout and commit at the end
		if err = vinDecodeNumber.Insert(ctx, txVinNumbers, boil.Infer()); err != nil {
			localLog.Err(err).
				Str("definition_id", tblDef.ID).
				Msg("failed to insert to vin_numbers")
		}
		err = txVinNumbers.Commit()
		if err != nil {
			return nil, errors.Wrap(err, "error when commiting transaction for inserting vin_number")
		}

		resp.DeviceMakeId = dm.ID //nolint
		resp.Manufacturer = dm.Name
		resp.Year = int32(vinDecodeNumber.Year)
		resp.Source = vinDecodeNumber.DecodeProvider.String
		pt, err := dc.powerTrainTypeService.ResolvePowerTrainType(split[0], split[1], null.JSON{}, null.JSON{})
		if err != nil {
			pt = coremodels.ICE.String()
		}
		resp.Powertrain = pt
		resp.DefinitionId = tblDef.ID

		metrics.Success.With(prometheus.Labels{"method": DeviceDefinitionOverride}).Inc()

		return resp, nil
	}

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
		return nil, errors.Wrap(err, "failed to get device_type")
	}
	// future: see if we can self decode model based on data we have before calling external decode WMI and VDS. Only thing is we won't get the style.

	if resp.Year == 0 || resp.Year > int32(time.Now().Year()+1) {
		localLog.Info().Msgf("encountered vin with non-standard year digit")
	}
	// check if this is a Tesla VIN, if not just follow regular path
	vinInfo := &coremodels.VINDecodingInfoData{}
	dbWMI, err := models.Wmis(models.WmiWhere.Wmi.EQ(wmi), qm.Load(models.WmiRels.DeviceMake)).One(ctx, dc.dbs().Reader)
	if err == nil && dbWMI != nil {
		if dbWMI.R.DeviceMake != nil && dbWMI.R.DeviceMake.Name == "Tesla" {
			vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.TeslaProvider, qry.Country)
			resp.Manufacturer = "Tesla"
			resp.DeviceMakeId = dbWMI.R.DeviceMake.ID //nolint
		}
	}
	// not a tesla, regular decode path
	if vinInfo == nil || vinInfo.Model == "" {
		vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.AllProviders, qry.Country) // this will try drivly first
	}

	// if no luck decoding VIN, try buildingVinInfo from known data passed in, typically smartcar or software connections
	if err != nil {
		if len(qry.KnownModel) > 0 && qry.KnownYear > 0 {
			// note if this is successful, err gets set to nil
			// todo: the knownModel should correspond with the Make
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
	// we may have already gotten this above
	if dbWMI == nil {
		dbWMI, err = dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
		if err != nil {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			dc.logger.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
			return resp, nil
		}
	}
	resp.DeviceMakeId = dbWMI.DeviceMakeID //nolint
	resp.Manufacturer = vinInfo.Make
	resp.Source = string(vinInfo.Source)
	resp.Year = vinInfo.Year
	resp.Model = vinInfo.Model

	modelSlug := shared.SlugString(vinInfo.Model)
	tid := common.DeviceDefinitionSlug(dbWMI.R.DeviceMake.NameSlug, modelSlug, int16(vinInfo.Year))
	tblDef, _, errTbl := dc.deviceDefinitionOnChainService.GetDefinitionByID(ctx, tid, dc.dbs().Reader)
	if errTbl != nil {
		dc.logger.Warn().Err(errTbl).Msgf("failed to get definition from tableland for vin: %s, id: %s", vin.String(), tid)
	} else if tblDef == nil {
		dc.logger.Warn().Msgf("failed to get definition from tableland for vin: %s, id: %s", vin.String(), tid)
	} else {
		dc.logger.Info().Str("vin", vin.String()).Msgf("found definition from tableland %s: %+v", tid, tblDef)
	}

	ddExists := true
	// if dd not found in tableland, we want to creat it
	if tblDef == nil {
		// metadata will only be set on create
		if errors.Is(err, sql.ErrNoRows) {
			// this method creates on-chain as well as in dd database
			dd, err := dc.ddRepository.GetOrCreate(ctx, txVinNumbers,
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
			localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", modelSlug)
			// set the response definition id from the newly created DD
			resp.DefinitionId = dd.NameSlug
		} else {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			return nil, err
		}
	} else {
		resp.DefinitionId = tblDef.ID // we can use the one from tableland since that exists
	}

	// match style - only process style if name is longer than 1
	if len(vinInfo.StyleName) < 2 {
		localLog.Warn().Msgf("decoded style name too short: %s must have a minimum of 2 characters.", vinInfo.StyleName)
	} else {
		externalStyleID := shared.SlugString(vinInfo.StyleName)
		// see if match existing style exists. First search is based on db device_definition_style_idx
		style, err := models.DeviceStyles(models.DeviceStyleWhere.DefinitionID.EQ(resp.DefinitionId),
			models.DeviceStyleWhere.Source.EQ(string(vinInfo.Source)),
			models.DeviceStyleWhere.ExternalStyleID.EQ(externalStyleID)).One(ctx, dc.dbs().Reader)
		if errors.Is(err, sql.ErrNoRows) {
			style, err = models.DeviceStyles(models.DeviceStyleWhere.DefinitionID.EQ(resp.DefinitionId),
				models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName)).One(ctx, dc.dbs().Reader)
		}

		if errors.Is(err, sql.ErrNoRows) {
			style = &models.DeviceStyle{
				ID:              ksuid.New().String(),
				DefinitionID:    resp.DefinitionId,
				Name:            vinInfo.StyleName,
				ExternalStyleID: externalStyleID,
				Source:          string(vinInfo.Source),
				SubModel:        vinInfo.SubModel,
				Metadata:        vinInfo.MetaData,
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
			// todo test error: device definition does not exist in the database
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
		Vin:              vin.String(),
		ManufacturerName: resp.Manufacturer,
		Wmi:              wmi,
		VDS:              vin.VDS(),
		Vis:              vin.VIS(),
		CheckDigit:       vin.CheckDigit(),
		SerialNumber:     vin.SerialNumber(),
		DecodeProvider:   null.StringFrom(string(vinInfo.Source)),
		Year:             int(resp.Year),
		DefinitionID:     resp.DefinitionId,
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

	localLog.Info().Str("device_definition_id", resp.DefinitionId).
		Str("style_id", resp.DeviceStyleId).
		Str("wmi", wmi).
		Str("vds", vin.VDS()).
		Str("vis", vin.VIS()).
		Str("check_digit", vin.CheckDigit()).Msgf("decoded vin ok with: %s", vinInfo.Source)
	// note we use a transaction here all throughout and commit at the end
	if err = vinDecodeNumber.Insert(ctx, txVinNumbers, boil.Infer()); err != nil {
		localLog.Err(err).
			Str("device_definition_id", resp.DefinitionId).
			Msg("failed to insert to vin_numbers")
	}
	err = txVinNumbers.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "error when commiting transaction for inserting vin_number")
	}

	// resolve images
	images, _ := models.Images(models.ImageWhere.DefinitionID.EQ(resp.DefinitionId)).All(ctx, dc.dbs().Reader)
	localLog.Debug().Msgf("Current Images : %d", len(images))

	if len(images) == 0 {
		err = dc.associateImagesToDeviceDefinition(ctx, resp.DefinitionId, vinInfo.Make, vinInfo.Model, int(resp.Year), 2, 2)
		if err != nil {
			localLog.Err(err).Send()
		}

		err = dc.associateImagesToDeviceDefinition(ctx, resp.DefinitionId, vinInfo.Make, vinInfo.Model, int(resp.Year), 2, 6)
		if err != nil {
			localLog.Err(err).Send()
		}
	}

	// why are we doing this if already have something similar above?
	if !ddExists {
		dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.NameSlug.EQ(resp.DefinitionId),
			qm.Load(models.DeviceDefinitionRels.DefinitionDeviceStyles),
			qm.Load(models.DeviceDefinitionRels.DeviceType),
			qm.Load(models.DeviceDefinitionRels.DeviceMake),
			qm.Load(models.DeviceDefinitionRels.DefinitionImages)).One(ctx, dc.dbs().Reader)
		if err != nil {
			return nil, errors.Wrap(err, "error when get dd for update powertraintype")
		}

		// set the dd metadata if nothing there, if fails just continue. this is needed in current setup
		if !gjson.GetBytes(dd.Metadata.JSON, dt.Metadatakey).Exists() {
			// todo - future: merge metadata properties. Also set style specific metadata - multiple places
			dd.Metadata = vinInfo.MetaData
			_, _ = dd.Update(ctx, dc.dbs().Writer, boil.Whitelist(models.DeviceDefinitionColumns.Metadata, models.DeviceDefinitionColumns.UpdatedAt))
			// todo- future: add powertrain - but this can be style specific - vincario gets us primary FuelType
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
							powerTrainTypeValue, err = dc.powerTrainTypeService.ResolvePowerTrainType(dd.R.DeviceMake.NameSlug, dd.ModelSlug, drivlyData, vincarioData)
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

		// todo this is already being done in above block
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
		split := strings.Split(resp.DefinitionId, "_")
		pt, _ := dc.powerTrainTypeService.ResolvePowerTrainType(split[0], split[1], vinDecodeNumber.DrivlyData, vinDecodeNumber.VincarioData)
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

func (dc DecodeVINQueryHandler) associateImagesToDeviceDefinition(ctx context.Context, definitionID, mk, model string, year int, prodID int, prodFormat int) error {

	img, err := dc.fuelAPIService.FetchDeviceImages(mk, model, year, prodID, prodFormat)
	if err != nil {
		dc.logger.Warn().Err(err).Msgf("unable to fetch device image for: %d %s %s", year, mk, model)
		return nil
	}

	var p models.Image

	// loop through all img (color variations)
	for _, device := range img.Images {
		p.ID = ksuid.New().String()
		p.DefinitionID = definitionID
		p.FuelAPIID = null.StringFrom(img.FuelAPIID)
		p.Width = null.IntFrom(img.Width)
		p.Height = null.IntFrom(img.Height)
		p.SourceURL = device.SourceURL
		//p.DimoS3URL = null.StringFrom("") // dont set it so it is null
		p.Color = device.Color
		p.NotExactImage = img.NotExactImage

		err = p.Upsert(ctx, dc.dbs().Writer, true, []string{models.ImageColumns.DefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
		if err != nil {
			dc.logger.Warn().Msgf("fail insert device image for: %s %d %s %s", definitionID, year, mk, model)
			continue
		}
	}

	return nil
}
