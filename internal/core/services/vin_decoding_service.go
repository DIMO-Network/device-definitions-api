//go:generate mockgen -source vin_decoding_service.go -destination mocks/vin_decoding_service_mock.go -package mocks

package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
)

type VINDecodingService interface {
	GetVIN(vin string, dt *repoModel.DeviceType, provider models.DecodeProviderEnum) (*models.VINDecodingInfoData, error)
}

type vinDecodingService struct {
	drivlyAPISvc   gateways.DrivlyAPIService
	vincarioAPISvc gateways.VincarioAPIService
	logger         *zerolog.Logger
}

func NewVINDecodingService(drivlyAPISvc gateways.DrivlyAPIService, vincarioAPISvc gateways.VincarioAPIService, logger *zerolog.Logger) VINDecodingService {
	return &vinDecodingService{drivlyAPISvc: drivlyAPISvc, vincarioAPISvc: vincarioAPISvc, logger: logger}
}

func (c vinDecodingService) GetVIN(vin string, dt *repoModel.DeviceType, provider models.DecodeProviderEnum) (*models.VINDecodingInfoData, error) {
	result := &models.VINDecodingInfoData{}
	vin = strings.ToUpper(strings.TrimSpace(vin))
	if !validateVIN(vin) {
		return nil, fmt.Errorf("invalid vin: %s", vin)
	}

	switch provider {
	case models.DrivlyProvider:
		vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with drivly", vin)
		}
		result = buildFromDrivly(vinDrivlyInfo)
	case models.VincarioProvider:
		vinVincarioInfo, err := c.vincarioAPISvc.DecodeVIN(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with vincario", vin)
		}
		result = buildFromVincario(vinVincarioInfo)
	case models.AllProviders:
		vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)
		if err != nil {
			c.logger.Warn().Err(err).Msg("could not decode vin with drivly")
		}
		if err == nil && vinDrivlyInfo != nil {
			if len(vinDrivlyInfo.Year) > 0 && len(vinDrivlyInfo.Make) > 0 && len(vinDrivlyInfo.Model) > 0 {
				result = buildFromDrivly(vinDrivlyInfo)
				metadata, err := common.BuildDeviceTypeAttributes(buildDrivlyVINInfoToUpdateAttr(vinDrivlyInfo), dt)
				if err != nil {
					c.logger.Warn().Err(err).Msg("unable to build metadata attributes")
				}
				result.MetaData = metadata
			}
		}
		// if nothing from drivly try vincario
		if result.Source == "" {
			vinVincarioInfo, err := c.vincarioAPISvc.DecodeVIN(vin)
			if err != nil {
				c.logger.Warn().Err(err).Msg("could not decode vin with vincario")
			}
			if err == nil && vinVincarioInfo != nil {
				result = buildFromVincario(vinVincarioInfo)
			}
		}
	}
	// could not decode anything
	if result.Source == "" {
		return nil, fmt.Errorf("could not decode from any provider for vin: %s", vin)
	}

	return result, nil
}

func validateVIN(vin string) bool {
	if len(vin) != 17 {
		return false
	}
	// match alpha numeric
	pattern := "[0-9A-Fa-f]+"
	regex := regexp.MustCompile(pattern)

	return regex.MatchString(vin)
}

func buildDrivlyVINInfoToUpdateAttr(vinInfo *gateways.DrivlyVINResponse) []*models.UpdateDeviceTypeAttribute {
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

func buildFromVincario(info *gateways.VincarioInfoResponse) *models.VINDecodingInfoData {
	raw, _ := json.Marshal(info)
	v := &models.VINDecodingInfoData{
		VIN:        info.VIN,
		Year:       strconv.Itoa(info.ModelYear),
		Make:       strings.TrimSpace(info.Make),
		Model:      strings.TrimSpace(info.Model),
		Source:     models.VincarioProvider,
		ExternalID: strconv.Itoa(info.VehicleID),
		StyleName:  info.GetStyle(),
		SubModel:   info.GetSubModel(),
		Raw:        raw,
	}
	return v
}

func buildFromDrivly(info *gateways.DrivlyVINResponse) *models.VINDecodingInfoData {
	raw, _ := json.Marshal(info)
	v := &models.VINDecodingInfoData{
		VIN:        info.Vin,
		Year:       info.Year,
		Make:       info.Make,
		Model:      info.Model,
		StyleName:  buildDrivlyStyleName(info),
		ExternalID: info.GetExternalID(),
		Source:     models.DrivlyProvider,
		Raw:        raw,
	}
	return v
}

func buildDrivlyStyleName(vinInfo *gateways.DrivlyVINResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}
