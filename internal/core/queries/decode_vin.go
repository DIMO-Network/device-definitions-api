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

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	"github.com/DIMO-Network/shared/pkg/vin"
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
	logger *zerolog.Logger,
	fuelAPIService gateways.FuelAPIService,
	powerTrainTypeService services.PowerTrainTypeService,
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		dbs:                            dbs,
		vinDecodingService:             vinDecodingService,
		logger:                         logger,
		vinRepository:                  vinRepository,
		fuelAPIService:                 fuelAPIService,
		powerTrainTypeService:          powerTrainTypeService,
		deviceDefinitionOnChainService: deviceDefinitionOnChainService,
	}
}

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*DecodeVINQuery)
	if len(qry.VIN) < 13 || len(qry.VIN) > 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
	resp := &p_grpc.DecodeVinResponse{}
	vin := vin.VIN(qry.VIN)
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
	// if database vin_number match found, just return it here
	if r := dc.hydrateResponseFromVinNumber(vinDecodeNumber); r != nil {
		metrics.Success.With(prometheus.Labels{"method": VinExists}).Inc()
		return r, nil
	}

	// todo: this should be a separate specific gRPC endpoint for setting or updating vin number to DD mapping
	// If DefinitionID passed in, override VIN decoding
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
			Wmi:              null.StringFrom(dbWMI.Wmi),
			VDS:              null.StringFrom(vin.VDS()),
			Vis:              null.StringFrom(vin.VIS()),
			CheckDigit:       null.StringFrom(vin.CheckDigit()),
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
	dbWMI, err := models.Wmis(models.WmiWhere.Wmi.EQ(wmi)).One(ctx, dc.dbs().Reader)
	if err == nil && dbWMI != nil {
		if dbWMI.ManufacturerName == "Tesla" {
			vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.TeslaProvider, qry.Country)
			resp.Manufacturer = "Tesla"
		}
	}
	// not a tesla, regular decode path
	if vinInfo == nil || vinInfo.Model == "" {
		vinInfo, err = dc.vinDecodingService.GetVIN(ctx, vin.String(), dt, coremodels.AllProviders, qry.Country) // this will try drivly first unless of japan
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
		// todo track failed decodes
		return nil, err
	}
	// WMI's may be re-used by multiple OEM's of same parent OEM, but just create it if needed
	if dbWMI == nil {
		_, err = dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
		if err != nil {
			// just log, Japan chasis numbers won't really work with this anyways
			dc.logger.Error().Err(err).Msgf("failed to get or create wmi for vin %s", vin.String())
		}
	}
	resp.Manufacturer = vinInfo.Make
	resp.Source = string(vinInfo.Source)
	resp.Year = vinInfo.Year
	resp.Model = vinInfo.Model

	modelSlug := stringutils.SlugString(vinInfo.Model)
	tid := common.DeviceDefinitionSlug(stringutils.SlugString(vinInfo.Make), modelSlug, int16(vinInfo.Year))
	resp.DefinitionId = tid

	tblDef, _, errTbl := dc.deviceDefinitionOnChainService.GetDefinitionByID(ctx, tid, dc.dbs().Reader)
	if errTbl != nil {
		dc.logger.Warn().Err(errTbl).Msgf("failed to get definition from tableland for vin: %s, id: %s", vin.String(), tid)
	} else if tblDef == nil {
		dc.logger.Warn().Msgf("failed to get definition from tableland for vin: %s, id: %s", vin.String(), tid)
	} else {
		dc.logger.Info().Str("vin", vin.String()).Msgf("found definition from tableland %s: %+v", tid, tblDef)
	}

	// add images if we don't have any for this definition_id
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

	// figure out powertrain
	pt := dc.powerTrainTypeService.ResolvePowerTrainFromVinInfo(vinInfo.StyleName, vinInfo.FuelType)
	if pt == "" {
		// try a different way
		pt, _ = dc.powerTrainTypeService.ResolvePowerTrainType(stringutils.SlugString(resp.Manufacturer), stringutils.SlugString(resp.Model), null.JSON{}, null.JSON{})
	}
	if pt != "" {
		resp.Powertrain = pt
	}

	// if dd not found in tableland, we want to create it
	if tblDef != nil {
		resp.DefinitionId = tblDef.ID
	} else {
		// if any images were added above, they will be in the database
		latestImages, _ := models.Images(models.ImageWhere.DefinitionID.EQ(resp.DefinitionId)).All(ctx, dc.dbs().Reader)
		// todo load up some metadata from what was decoded. Powertrain too
		md := resolveMetadataFromInfo(resp.Powertrain, vinInfo)

		trx, err := dc.deviceDefinitionOnChainService.Create(ctx, resp.Manufacturer, coremodels.DeviceDefinitionTablelandModel{
			ID:         tid,
			KSUID:      ksuid.New().String(),
			Model:      resp.Model,
			Year:       int(resp.Year),
			DeviceType: common.DefaultDeviceType,
			ImageURI:   common.GetDefaultImageURL(latestImages),
			Metadata:   md,
		})
		if err != nil {
			metrics.InternalError.With(prometheus.Labels{"method": VinErrors}).Inc()
			return nil, errors.Wrap(err, "error creating new device definition on-chain from decoded vin")
		}
		resp.NewTrxHash = *trx
	}

	// match style - only process style if name is longer than 1
	if len(vinInfo.StyleName) < 2 {
		localLog.Warn().Msgf("decoded style name too short: %s must have a minimum of 2 characters.", vinInfo.StyleName)
	} else {
		var styleErr error
		resp.DeviceStyleId, styleErr = dc.processDeviceStyle(ctx, vinInfo, tid, resp.Powertrain)
		if styleErr != nil {
			dc.logger.Error().Err(styleErr).Msgf("error processing device style for vin: %s. continuing", vin.String())
		}
	}

	// insert vin_numbers
	errVinNumber := dc.saveVinDecodeNumber(ctx, vin, vinInfo, resp)
	if errVinNumber != nil {
		return nil, errors.Wrap(errVinNumber, "error saving vin_number")
	}

	localLog.Info().Str("device_definition_id", resp.DefinitionId).
		Str("style_id", resp.DeviceStyleId).
		Str("wmi", wmi).
		Str("vds", vin.VDS()).
		Str("vis", vin.VIS()).
		Str("check_digit", vin.CheckDigit()).Msgf("decoded vin ok with: %s", vinInfo.Source)

	metrics.Success.With(prometheus.Labels{"method": VinSuccess}).Inc()

	return resp, nil
}

func resolveMetadataFromInfo(powertrain string, _ *coremodels.VINDecodingInfoData) *coremodels.DeviceDefinitionMetadata {
	md := coremodels.DeviceDefinitionMetadata{DeviceAttributes: make([]coremodels.DeviceTypeAttribute, 0)}
	if powertrain != "" {
		md.DeviceAttributes = append(md.DeviceAttributes, coremodels.DeviceTypeAttribute{
			Name:  common.PowerTrainType,
			Value: powertrain,
		})
	}

	return &md
}

// hydrateResponseFromVinNumber pass in a vin_number database object and converts to vin decode response
func (dc DecodeVINQueryHandler) hydrateResponseFromVinNumber(vn *models.VinNumber) *p_grpc.DecodeVinResponse {
	if vn == nil {
		return nil
	}
	// call on-chain svc to get the DD and pull out the powertrain
	pt := ""
	trx := ""
	tblDef, manufID, err := dc.deviceDefinitionOnChainService.GetDefinitionByID(context.Background(), vn.DefinitionID, dc.dbs().Reader)
	if err == nil && tblDef != nil {
		for _, attribute := range tblDef.Metadata.DeviceAttributes {
			if attribute.Name == common.PowerTrainType {
				pt = attribute.Value
			}
		}
		if pt == "" {
			makeName, _ := dc.deviceDefinitionOnChainService.GetManufacturerNameByID(context.Background(), manufID)
			pt, _ = dc.powerTrainTypeService.ResolvePowerTrainType(stringutils.SlugString(makeName), stringutils.SlugString(tblDef.Model), null.JSON{}, null.JSON{})
		}
	} else {
		// this is not good, somehow it got decoded in past without it being created on tableland
		dc.logger.Warn().Msgf("vin decoded for unexistent device definition: %s, vin: %s", vn.DefinitionID, vn.Vin)
	}

	resp := &p_grpc.DecodeVinResponse{
		Manufacturer:  vn.ManufacturerName,
		Year:          int32(vn.Year),
		DeviceStyleId: vn.StyleID.String,
		Source:        vn.DecodeProvider.String,
		DefinitionId:  vn.DefinitionID,
		Powertrain:    pt,
		NewTrxHash:    trx,
	}

	return resp
}

// processDeviceStyle saves new styles if needed to db and returns the style database ID
func (dc DecodeVINQueryHandler) processDeviceStyle(ctx context.Context, vinInfo *coremodels.VINDecodingInfoData, definitionID, powertrain string) (string, error) {
	externalStyleID := stringutils.SlugString(vinInfo.StyleName)

	// Step 1: Try to find an existing style
	style, err := models.DeviceStyles(
		models.DeviceStyleWhere.DefinitionID.EQ(definitionID),
		models.DeviceStyleWhere.Source.EQ(string(vinInfo.Source)),
		models.DeviceStyleWhere.ExternalStyleID.EQ(externalStyleID),
	).One(ctx, dc.dbs().Reader)

	if errors.Is(err, sql.ErrNoRows) {
		// Step 2: If not found, try searching by name
		style, err = models.DeviceStyles(
			models.DeviceStyleWhere.DefinitionID.EQ(definitionID),
			models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName),
		).One(ctx, dc.dbs().Reader)
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Step 3: Create a new style if it doesn't exist
		style = &models.DeviceStyle{
			ID:              ksuid.New().String(),
			DefinitionID:    definitionID,
			Name:            vinInfo.StyleName,
			ExternalStyleID: externalStyleID,
			Source:          string(vinInfo.Source),
			SubModel:        vinInfo.SubModel,
			Metadata:        vinInfo.MetaData,
		}

		// Resolve powertrain and add to metadata if applicable
		if powertrain != "" {
			metadataWithPT, metadataErr := sjson.SetBytes(vinInfo.MetaData.JSON, common.PowerTrainType, powertrain)
			if metadataErr == nil {
				style.Metadata = null.JSONFrom(metadataWithPT)
			}
		}

		// Insert the new style into the database
		errStyle := style.Insert(ctx, dc.dbs().Writer, boil.Infer())
		if errStyle != nil {
			return "", errors.Wrapf(errStyle, "error creating style with values: %+v", style)
		}
	}
	return style.ID, nil
}

func (dc DecodeVINQueryHandler) saveVinDecodeNumber(ctx context.Context, vin vin.VIN, vinInfo *coremodels.VINDecodingInfoData, resp *p_grpc.DecodeVinResponse) error {
	vinDecodeNumber := &models.VinNumber{
		Vin:              vin.String(),
		ManufacturerName: resp.Manufacturer,
		Wmi:              null.StringFrom(vin.Wmi()),
		SerialNumber:     vin.SerialNumber(),
		DecodeProvider:   null.StringFrom(string(vinInfo.Source)),
		Year:             int(resp.Year),
		DefinitionID:     resp.DefinitionId,
	}
	if !vin.IsJapanChassis() {
		vinDecodeNumber.VDS = null.StringFrom(vin.VDS())
		vinDecodeNumber.Vis = null.StringFrom(vin.VIS())
		vinDecodeNumber.CheckDigit = null.StringFrom(vin.CheckDigit())
	}

	// Optional fields based on response and VIN info
	if len(resp.DeviceStyleId) > 0 {
		vinDecodeNumber.StyleID = null.StringFrom(resp.DeviceStyleId)
	}

	switch vinInfo.Source {
	case coremodels.DrivlyProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.DrivlyData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.VincarioProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.VincarioData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.AutoIsoProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.AutoisoData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.DATGroupProvider:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.DatgroupData = null.JSONFrom(vinInfo.Raw)
		}
	case coremodels.Japan17VIN:
		if len(vinInfo.Raw) > 0 {
			vinDecodeNumber.Vin17Data = null.JSONFrom(vinInfo.Raw)
		}
	}

	// Insert VIN decode number into the database
	if err := vinDecodeNumber.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
		return errors.Wrapf(err, "error inserting vin_number with values: %+v", vinDecodeNumber)
	}
	return nil
}

// vinInfoFromKnown builds a vininfo object based on one passed in with Make from vin WMI, and passed in model and year set
func (dc DecodeVINQueryHandler) vinInfoFromKnown(vin vin.VIN, knownModel string, knownYear int32) (*coremodels.VINDecodingInfoData, error) {
	vinInfo := &coremodels.VINDecodingInfoData{}
	vinInfo.VIN = vin.String()
	wmis, err := models.Wmis(models.WmiWhere.Wmi.EQ(vin.Wmi())).All(context.Background(), dc.dbs().Reader)
	if err != nil {
		return nil, errors.Wrap(err, "vinInfoFromKnown: could not get WMI from vin wmi "+vin.Wmi())
	}
	if len(wmis) > 1 {
		return nil, fmt.Errorf("vinInfoFromKnown: more than one WMI found for vin wmi %s", vin.Wmi())
	}
	vinInfo.Make = wmis[0].ManufacturerName
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
			dc.logger.Warn().Err(err).Msgf("fail insert device image for: %s %d %s %s", definitionID, year, mk, model)
			continue
		}
	}

	return nil
}
