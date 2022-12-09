package queries

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DecodeVINQueryHandler struct {
	dbs          func() *db.ReaderWriter
	drivlyAPISvc gateways.DrivlyAPIService
	logger       *zerolog.Logger
	repository   repositories.DeviceDefinitionRepository
}

type DecodeVINQuery struct {
	VIN string `json:"vin"`
}

func (*DecodeVINQuery) Key() string { return "DecodeVINQuery" }

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter, drivlyAPISvc gateways.DrivlyAPIService, repository repositories.DeviceDefinitionRepository, logger *zerolog.Logger) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		dbs:          dbs,
		drivlyAPISvc: drivlyAPISvc,
		logger:       logger,
		repository:   repository,
	}
}

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*DecodeVINQuery)
	if len(qry.VIN) != 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
	resp := &p_grpc.DecodeVINResponse{}
	// get the year
	vin := shared.VIN(qry.VIN)
	resp.Year = int32(vin.Year()) // needs to be updated for newer years
	// todo: we could decode tesla on our own
	// get the make
	wmi := qry.VIN[0:3]
	dbWMI, err := models.FindWmi(ctx, dc.dbs().Reader, wmi)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if dbWMI != nil {
		resp.DeviceMakeId = dbWMI.DeviceMakeID
	}

	localLog := dc.logger.With().Str("vin", qry.VIN).Str("handler", qry.Key()).Str("year", string(resp.Year)).Logger()
	// not yet - lookup the device definition by rest of info - look at our existing vins
	// for now always call drivly to decode
	vinInfo, err := dc.drivlyAPISvc.GetVINInfo(vin.String())
	if err != nil {
		localLog.Err(err).Msg("failed to decode vin from drivly")
		return resp, nil
	}
	// get the make from the vinInfo if no WMI found
	if dbWMI == nil {
		deviceMake, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(common.SlugString(vinInfo.Make))).One(ctx, dc.dbs().Reader)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				localLog.Warn().Msgf("failed to find make from vin decode with name slug: %s", common.SlugString(vinInfo.Make))
			} else {
				return nil, err
			}
		}
		if deviceMake != nil {
			resp.DeviceMakeId = deviceMake.ID
			// insert the WMI
			dbWMI = &models.Wmi{
				Wmi:          wmi,
				DeviceMakeID: deviceMake.ID,
			}
			if err = dbWMI.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
				localLog.Err(err).Str("deviceMakeId", deviceMake.ID).Msgf("failed to insert wmi: %s", wmi)
			}
		}
	}
	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, dc.dbs().Reader)
	if err != nil {
		return nil, err
	}
	metadata, err := common.BuildDeviceTypeAttributes(drivlyVINInfoToUpdateAttr(vinInfo), dt)
	if err != nil {
		localLog.Err(err).Msg("unable to build metadata attributes")
	}
	// now match the model for the dd id
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(resp.DeviceMakeId),
		models.DeviceDefinitionWhere.Year.EQ(int16(resp.Year)),
		models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
		One(ctx, dc.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dd, err = dc.repository.GetOrCreate(ctx, "drivly", common.SlugString(vinInfo.Model+vinInfo.Year), resp.DeviceMakeId,
				vinInfo.Model, int(resp.Year), common.DefaultDeviceType, metadata, true)
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
			models.DeviceStyleWhere.SubModel.EQ(vinInfo.SubModel),
			models.DeviceStyleWhere.Name.EQ(buildStyleName(vinInfo))).One(ctx, dc.dbs().Reader)
		if errors.Is(err, sql.ErrNoRows) {
			// insert
			style = &models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: dd.ID,
				Name:               buildStyleName(vinInfo),
				ExternalStyleID:    common.SlugString(buildStyleName(vinInfo)),
				Source:             "drivly",
				SubModel:           vinInfo.SubModel,
			}
			_ = style.Insert(ctx, dc.dbs().Writer, boil.Infer())
		}
		resp.DeviceStyleId = style.ID
		// set the dd metadata if nothing there
		if !gjson.GetBytes(dd.Metadata.JSON, dt.Metadatakey).Exists() {
			// todo - future: merge properties as needed. Also set style specific metadata - multiple places
			dd.Metadata = metadata
			_, _ = dd.Update(ctx, dc.dbs().Writer, boil.Whitelist(models.DeviceDefinitionColumns.Metadata, models.DeviceDefinitionColumns.UpdatedAt))
		}
		// todo- future: add powertrain - but this can be style specific
	}

	return resp, nil
}

func buildStyleName(vinInfo *gateways.VINInfoResponse) string {
	return vinInfo.Trim + " " + vinInfo.SubModel
}

func drivlyVINInfoToUpdateAttr(vinInfo *gateways.VINInfoResponse) []*coremodels.UpdateDeviceTypeAttribute {
	seekAttributes := map[string]string{
		// {device attribute, must match device_types.properties}: {vin info from drivly}
		"mpg_city":               "mpgCity",
		"mpg_highway":            "mpgHighway",
		"mpg":                    "mpg",
		"base_msrp":              "msrpBase",
		"fuel_tank_capacity_gal": "fuelTankCapacityGal",
		"fuel_type":              "fuel",
		"wheelbase":              "wheelbase",
		"generation":             "generation",
		"number_of_doors":        "doors",
		"manufacturer_code":      "manufacturerCode",
		"driven_wheels":          "drive",
	}
	marshal, _ := json.Marshal(vinInfo)
	var udta []*coremodels.UpdateDeviceTypeAttribute

	for dtAttrKey, drivlyKey := range seekAttributes {
		v := gjson.GetBytes(marshal, drivlyKey).String()
		// if v valid, ok etc
		if len(v) > 0 && v != "0" && v != "0.0000" {
			udta = append(udta, &coremodels.UpdateDeviceTypeAttribute{
				Name:  dtAttrKey,
				Value: v,
			})
		}
	}

	return udta
}
