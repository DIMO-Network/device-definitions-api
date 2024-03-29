//go:generate mockgen -source vin_decoding_service.go -destination mocks/vin_decoding_service_mock.go -package mocks

package services

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
)

type VINDecodingService interface {
	GetVIN(ctx context.Context, vin string, dt *repoModel.DeviceType, provider models.DecodeProviderEnum, country string) (*models.VINDecodingInfoData, error)
}

type vinDecodingService struct {
	drivlyAPISvc       gateways.DrivlyAPIService
	vincarioAPISvc     gateways.VincarioAPIService
	logger             *zerolog.Logger
	repository         repositories.DeviceDefinitionRepository
	autoIsoAPIService  gateways.AutoIsoAPIService
	DATGroupAPIService gateways.DATGroupAPIService
}

func NewVINDecodingService(drivlyAPISvc gateways.DrivlyAPIService, vincarioAPISvc gateways.VincarioAPIService, autoIso gateways.AutoIsoAPIService, logger *zerolog.Logger, repository repositories.DeviceDefinitionRepository, datGroupAPIService gateways.DATGroupAPIService) VINDecodingService {
	return &vinDecodingService{drivlyAPISvc: drivlyAPISvc, vincarioAPISvc: vincarioAPISvc, autoIsoAPIService: autoIso, logger: logger, repository: repository, DATGroupAPIService: datGroupAPIService}
}

func (c vinDecodingService) GetVIN(ctx context.Context, vin string, dt *repoModel.DeviceType, provider models.DecodeProviderEnum, country string) (*models.VINDecodingInfoData, error) {

	const DefaultDeviceDefinitionID = "22N2y6TCaDBYPUsXJb3u02bqN2I"

	result := &models.VINDecodingInfoData{}
	vin = strings.ToUpper(strings.TrimSpace(vin))
	if !validateVIN(vin) {
		return nil, fmt.Errorf("invalid vin: %s", vin)
	}

	localLog := c.logger.With().
		Str("vin", vin).
		Logger()

	if strings.HasPrefix(vin, "0SC") {
		dd, err := c.repository.GetByID(ctx, DefaultDeviceDefinitionID)
		if err != nil {
			return nil, err
		}
		result = buildFromDD(vin, dd)
		return result, nil
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
		result, err = buildFromVincario(vinVincarioInfo)
		if err != nil {
			return nil, err
		}
	case models.AutoIsoProvider:
		vinAutoIsoInfo, err := c.autoIsoAPIService.GetVIN(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with autoiso", vin)
		}
		result, err = buildFromAutoIso(vinAutoIsoInfo)
		if err != nil {
			return nil, err
		}
	case models.DATGroupProvider:
		vinInfo, err := c.DATGroupAPIService.GetVIN(vin, country)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with DATGroup", vin)
		}
		result, err = buildFromDATGroup(vinInfo)
		if err != nil {
			return nil, err
		}
	case models.AllProviders:
		vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)
		if err != nil {
			localLog.Warn().Err(err).Msg("AllProviders decode - unable decode vin with drivly")
		} else {
			result = buildFromDrivly(vinDrivlyInfo)
			metadata, err := common.BuildDeviceTypeAttributes(buildDrivlyVINInfoToUpdateAttr(vinDrivlyInfo), dt)
			if err != nil {
				localLog.Warn().Err(err).Msg("AllProviders decode - unable to build metadata attributes")
			}
			result.MetaData = metadata
		}
		// if nothing from drivly, try DATGroup
		if result.Source == "" {
			datGroupInfo, err := c.DATGroupAPIService.GetVIN(vin, country)
			if err != nil {
				localLog.Warn().Err(err).Msg("AllProviders decode -could not decode vin with DATGroup")
			} else if datGroupInfo == nil {
				localLog.Warn().Msg("datgroup returned nil object, but no error")
			} else {
				result, err = buildFromDATGroup(datGroupInfo)
				if err != nil {
					localLog.Warn().Err(err).Msg("AllProviders decode -could not build struct from DATGroup data")
				}
			}
		}
		// if nothing from drivly, try autoiso
		if result.Source == "" {
			autoIsoInfo, err := c.autoIsoAPIService.GetVIN(vin)
			if err != nil {
				localLog.Warn().Err(err).Msg("AllProviders decode -could not decode vin with autoiso")
			} else {
				result, err = buildFromAutoIso(autoIsoInfo)
				if err != nil {
					localLog.Warn().Err(err).Msg("AllProviders decode -could not build struct from autoiso data")
				}
			}
		}
		// if nothing from autoiso try vincario
		if result.Source == "" {
			vinVincarioInfo, err := c.vincarioAPISvc.DecodeVIN(vin)
			if err != nil {
				localLog.Warn().Err(err).Msg("AllProviders decode -could not decode vin with vincario")
			} else {
				result, err = buildFromVincario(vinVincarioInfo)
				if err != nil {
					localLog.Warn().Err(err).Msg("AllProviders decode -could not build struct from vincario data")
				}
			}
		}
	}
	// could not decode anything
	if result.Source == "" {
		return nil, fmt.Errorf("could not decode from any provider for vin: %s", vin)
	}
	if result.Year == 0 {
		return nil, fmt.Errorf("unable to decode vin: %s - year returned as 0", vin)
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

func buildFromAutoIso(info *gateways.AutoIsoVINResponse) (*models.VINDecodingInfoData, error) {
	raw, _ := json.Marshal(info)
	if info == nil {
		return nil, fmt.Errorf("vin info was nil")
	}
	yr, err := strconv.Atoi(info.FunctionResponse.Data.Decoder.ModelYear.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid decode year: %+v", *info)
	}
	v := &models.VINDecodingInfoData{
		VIN:        info.Vin,
		Year:       int32(yr),
		Make:       strings.TrimSpace(info.FunctionResponse.Data.Decoder.Make.Value),
		Model:      strings.TrimSpace(info.FunctionResponse.Data.Decoder.Model.Value),
		Source:     models.AutoIsoProvider,
		ExternalID: info.Vin,
		StyleName:  info.GetStyle(),
		SubModel:   info.GetSubModel(),
		Raw:        raw,
		FuelType:   info.FunctionResponse.Data.Decoder.FuelType.Value,
	}

	return v, nil
}

func buildFromVincario(info *gateways.VincarioInfoResponse) (*models.VINDecodingInfoData, error) {
	raw, _ := json.Marshal(info)
	v := &models.VINDecodingInfoData{
		VIN:        info.VIN,
		Year:       int32(info.ModelYear),
		Make:       strings.TrimSpace(info.Make),
		Model:      strings.TrimSpace(info.Model),
		Source:     models.VincarioProvider,
		ExternalID: strconv.Itoa(info.VehicleID),
		StyleName:  info.GetStyle(),
		SubModel:   info.GetSubModel(),
		Raw:        raw,
		FuelType:   info.FuelType,
	}
	m, err := info.GetMetadata()
	if err != nil {
		return nil, err
	}
	v.MetaData = m

	return v, nil
}

func buildFromDrivly(info *gateways.DrivlyVINResponse) *models.VINDecodingInfoData {
	raw, _ := json.Marshal(info)
	yrInt, _ := strconv.Atoi(info.Year)

	v := &models.VINDecodingInfoData{
		VIN:        info.Vin,
		Year:       int32(yrInt),
		Make:       info.Make,
		Model:      info.Model,
		StyleName:  buildDrivlyStyleName(info),
		ExternalID: info.GetExternalID(),
		Source:     models.DrivlyProvider,
		Raw:        raw,
		FuelType:   info.Fuel,
	}
	return v
}

func buildDrivlyStyleName(vinInfo *gateways.DrivlyVINResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}

func buildFromDD(vin string, info *repoModel.DeviceDefinition) *models.VINDecodingInfoData {

	v := &models.VINDecodingInfoData{
		VIN:   vin,
		Year:  int32(info.Year),
		Make:  info.R.DeviceMake.Name,
		Model: info.Model,
	}

	if len(info.R.DeviceStyles) > 0 {
		v.StyleName = info.R.DeviceStyles[0].Name
	}

	if info.ExternalID.Valid {
		v.ExternalID = info.ExternalID.String
	}

	return v
}

func buildFromDATGroup(info *gateways.GetVehicleIdentificationByVinResponse) (*models.VINDecodingInfoData, error) {
	raw, _ := xml.Marshal(info)
	if info == nil {
		return nil, fmt.Errorf("vin info was nil")
	}

	if len(info.Body.GetDataVehicleIdentificationByVinResponse.VXS.Dossier) == 0 {
		return nil, fmt.Errorf("no dosier resp from datgroup: %s", string(raw))
	}

	dossier := info.Body.GetDataVehicleIdentificationByVinResponse.VXS.Dossier[0].Vehicle
	v := &models.VINDecodingInfoData{
		VIN:        dossier.VINResult.VINVehicle.VINumber.VinCode,
		Year:       int32(dossier.BuildYear),
		Make:       strings.TrimSpace(dossier.ManufacturerName),
		Model:      strings.TrimSpace(dossier.BaseModelName),
		Source:     models.DATGroupProvider,
		ExternalID: dossier.VINResult.VINVehicle.VINumber.VinCode,
		StyleName:  dossier.SubModelName,
		//SubModel:   info.GetSubModel(),
		Raw: raw,
	}

	if dossier.TechInfo != nil {
		v.FuelType = dossier.TechInfo.FuelMethodType
	}

	return v, nil
}
