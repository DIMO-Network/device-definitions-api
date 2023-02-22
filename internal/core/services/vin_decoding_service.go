//go:generate mockgen -source vin_deconding_service.go -destination mocks/vin_deconding_service_mock.go -package mocks

package services

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
)

type VINDecodingService interface {
	GetVIN(vin string, dt *repoModel.DeviceType) (*models.VINDecodingInfoData, error)
}

type vinDecodingService struct {
	drivlyAPISvc   gateways.DrivlyAPIService
	vincarioAPISvc gateways.VincarioAPIService
	logger         *zerolog.Logger
}

func NewVINDecodingService(drivlyAPISvc gateways.DrivlyAPIService, vincarioAPISvc gateways.VincarioAPIService, logger *zerolog.Logger) VINDecodingService {
	return &vinDecodingService{drivlyAPISvc: drivlyAPISvc, vincarioAPISvc: vincarioAPISvc, logger: logger}
}

func (c vinDecodingService) GetVIN(vin string, dt *repoModel.DeviceType) (*models.VINDecodingInfoData, error) {
	vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)

	result := &models.VINDecodingInfoData{}

	if err != nil {
		c.logger.Debug().
			Str("vin", vin).
			Msg("failed to decode vin from drivly")
	}

	if vinDrivlyInfo == nil {
		vinVincarioInfo, err := c.vincarioAPISvc.DecodeVIN(vin)

		if err != nil {
			c.logger.Debug().
				Str("vin", vin).
				Msg("failed to decode vin from vincario")
		}

		if vinVincarioInfo != nil {
			result.VIN = vinVincarioInfo.VIN
			result.Year = strconv.Itoa(vinVincarioInfo.ModelYear)
			result.Make = vinVincarioInfo.Make
			result.Model = vinVincarioInfo.Model
			result.Source = "vincario"

			return result, nil
		}

		return nil, nil
	}

	result.VIN = vinDrivlyInfo.Vin
	result.Year = vinDrivlyInfo.Year
	result.Make = vinDrivlyInfo.Make
	result.Model = vinDrivlyInfo.Model
	result.StyleName = buildDrivlyStyleName(vinDrivlyInfo)
	result.Source = "drivly"

	metadata, err := common.BuildDeviceTypeAttributes(buildDrivlyVINInfoToUpdateAttr(vinDrivlyInfo), dt)
	if err != nil {
		c.logger.Warn().Err(err).Msg("unable to build metadata attributes")
	}

	result.MetaData = metadata

	return result, nil
}

func buildDrivlyVINInfoToUpdateAttr(vinInfo *gateways.VINInfoResponse) []*models.UpdateDeviceTypeAttribute {
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
	var udta []*models.UpdateDeviceTypeAttribute

	for dtAttrKey, drivlyKey := range seekAttributes {
		v := gjson.GetBytes(marshal, drivlyKey).String()
		// if v valid, ok etc
		if len(v) > 0 && v != "0" && v != "0.0000" {
			udta = append(udta, &models.UpdateDeviceTypeAttribute{
				Name:  dtAttrKey,
				Value: v,
			})
		}
	}

	return udta
}

func buildDrivlyStyleName(vinInfo *gateways.VINInfoResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}
