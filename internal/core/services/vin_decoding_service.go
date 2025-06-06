//go:generate mockgen -source vin_decoding_service.go -destination mocks/vin_decoding_service_mock.go -package mocks

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/shared/pkg/db"

	vinutil "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/volatiletech/null/v8"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
)

type VINDecodingService interface {
	// GetVIN decodes a vin using one of the providers passed in or if AllProviders applies an ordered logic. Only pass TeslaProvider if know it is a Tesla.
	GetVIN(ctx context.Context, vin string, dt *repoModel.DeviceType, provider coremodels.DecodeProviderEnum, country string) (*coremodels.VINDecodingInfoData, error)
}

type vinDecodingService struct {
	logger             *zerolog.Logger
	drivlyAPISvc       gateways.DrivlyAPIService
	vincarioAPISvc     gateways.VincarioAPIService
	autoIsoAPIService  gateways.AutoIsoAPIService
	DATGroupAPIService gateways.DATGroupAPIService
	japan17VINAPI      gateways.Japan17VINAPI
	onChainSvc         gateways.DeviceDefinitionOnChainService
	dbs                func() *db.ReaderWriter
}

func NewVINDecodingService(drivlyAPISvc gateways.DrivlyAPIService, vincarioAPISvc gateways.VincarioAPIService, autoIso gateways.AutoIsoAPIService, logger *zerolog.Logger,
	onChainSvc gateways.DeviceDefinitionOnChainService, datGroupAPIService gateways.DATGroupAPIService, dbs func() *db.ReaderWriter,
	japan17VINAPI gateways.Japan17VINAPI) VINDecodingService {
	return &vinDecodingService{drivlyAPISvc: drivlyAPISvc, vincarioAPISvc: vincarioAPISvc, autoIsoAPIService: autoIso,
		japan17VINAPI: japan17VINAPI, logger: logger, onChainSvc: onChainSvc, DATGroupAPIService: datGroupAPIService, dbs: dbs}
}

func (c vinDecodingService) GetVIN(ctx context.Context, vin string, dt *repoModel.DeviceType, provider coremodels.DecodeProviderEnum, country string) (*coremodels.VINDecodingInfoData, error) {

	const DefaultDefinitionID = "ford_escape_2020"

	result := &coremodels.VINDecodingInfoData{}
	vin = strings.ToUpper(strings.TrimSpace(vin))
	// check for japan chasis
	if (len(vin) < 17 && len(vin) > 10) || country == "JPN" {
		provider = coremodels.Japan17VIN
	} else if !ValidateVIN(vin) {
		return nil, fmt.Errorf("invalid vin: %s", vin)
	}

	localLog := c.logger.With().
		Str("vin", vin).
		Logger()

	if strings.HasPrefix(vin, "0SC") {
		dd, _, err := c.onChainSvc.GetDefinitionByID(ctx, DefaultDefinitionID)
		if err != nil {
			return nil, err
		}
		result = buildFromDDForTestVIN(vin, dd)
		return result, nil
	}

	switch provider {
	case coremodels.TeslaProvider:
		v := vinutil.VIN(vin)
		metadata := map[string]interface{}{
			"fuel_type":       "electric",
			"powertrain_type": coremodels.BEV.String(),
		}
		bytes, _ := json.Marshal(metadata)
		result = &coremodels.VINDecodingInfoData{
			VIN:      vin,
			Year:     int32(v.Year()),
			Make:     "Tesla",
			Model:    v.TeslaModel(),
			Source:   coremodels.TeslaProvider,
			FuelType: "electric",
			MetaData: null.JSONFrom(bytes),
		}
	case coremodels.DrivlyProvider:
		vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with drivly", vin)
		}
		result, err = buildFromDrivly(vinDrivlyInfo)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with drivly", vin)
		}
	case coremodels.VincarioProvider:
		vinVincarioInfo, err := c.vincarioAPISvc.DecodeVIN(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with vincario", vin)
		}
		result, err = buildFromVincario(vinVincarioInfo)
		if err != nil {
			return nil, err
		}
	case coremodels.Japan17VIN:
		mmy, raw, err := c.japan17VINAPI.GetVINInfo(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with japan17vin", vin)
		}
		result = &coremodels.VINDecodingInfoData{
			VIN:      vin,
			Make:     mmy.ManufacturerName,
			Model:    mmy.ModelName,
			Year:     int32(mmy.Year),
			Source:   coremodels.Japan17VIN,
			MetaData: null.JSONFrom(raw),
			Raw:      raw,
		}
	case coremodels.AutoIsoProvider:
		vinAutoIsoInfo, err := c.autoIsoAPIService.GetVIN(vin)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with autoiso", vin)
		}
		result, err = buildFromAutoIso(vinAutoIsoInfo)
		if err != nil {
			return nil, err
		}
	case coremodels.DATGroupProvider:
		// todo lookup country for two letter equiv
		vinInfo, err := c.DATGroupAPIService.GetVINv2(vin, country) // try with Turkey
		if err != nil {
			return nil, errors.Wrapf(err, "unable to decode vin: %s with DATGroup", vin)
		}
		result, err = buildFromDATGroup(vinInfo)
		if err != nil {
			return nil, err
		}
	case coremodels.AllProviders:
		// todo if tesla, just build from tesla and use model

		vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)
		if err != nil {
			localLog.Warn().Err(err).Msg("AllProviders decode - unable decode vin with drivly")
		} else {
			result, err = buildFromDrivly(vinDrivlyInfo)
			if err != nil {
				localLog.Warn().Err(err).Msg("AllProviders decode -could not decode vin with drivly")
			} else {
				metadata, err := common.BuildDeviceTypeAttributes(buildDrivlyVINInfoToUpdateAttr(vinDrivlyInfo), dt)
				if err != nil {
					localLog.Warn().Err(err).Msg("AllProviders decode - unable to build metadata attributes")
				}
				result.MetaData = metadata
			}
		}

		// if nothing from drivly, try autoiso
		if result == nil || result.Source == "" {
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

		// if nothing,try vincario
		if result == nil || result.Source == "" {
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

		// if nothing from vincario, try DATGroup
		if result == nil || result.Source == "" {
			// idea: only accept WMI's for DATgroup that they have succesfully decoded in the past
			datGroupInfo, err := c.DATGroupAPIService.GetVINv2(vin, country)
			if err != nil {
				localLog.Warn().Err(err).Msg("AllProviders decode -could not decode vin with DATGroup")
			} else {
				result, err = buildFromDATGroup(datGroupInfo)
				localLog.Info().Msgf("datgroup result: %+v", result) // temporary for debugging
				if err != nil {
					localLog.Warn().Err(err).Msg("AllProviders decode - could not build struct from DATGroup data")
				}
			}
		}
	}
	// could not decode anything
	if result == nil || result.Source == "" {
		return nil, fmt.Errorf("could not decode from any provider for vin: %s", vin)
	}
	if result.Year == 0 {
		return nil, fmt.Errorf("unable to decode vin: %s - year returned as 0", vin)
	}

	return result, nil
}

func ValidateVIN(vin string) bool {
	if len(vin) != 17 {
		return false
	}
	// match alpha numeric
	pattern := "[0-9A-Fa-f]+"
	regex := regexp.MustCompile(pattern)

	return regex.MatchString(vin)
}

func buildDrivlyVINInfoToUpdateAttr(vinInfo *coremodels.DrivlyVINResponse) []*coremodels.UpdateDeviceTypeAttribute {
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

func buildFromAutoIso(info *coremodels.AutoIsoVINResponse) (*coremodels.VINDecodingInfoData, error) {
	raw, _ := json.Marshal(info)
	if info == nil {
		return nil, fmt.Errorf("vin info was nil")
	}
	yr, err := strconv.Atoi(info.FunctionResponse.Data.Decoder.ModelYear.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid decode year: %+v", *info)
	}
	v := &coremodels.VINDecodingInfoData{
		VIN:        info.Vin,
		Year:       int32(yr),
		Make:       strings.TrimSpace(info.FunctionResponse.Data.Decoder.Make.Value),
		Model:      strings.TrimSpace(info.FunctionResponse.Data.Decoder.Model.Value),
		Source:     coremodels.AutoIsoProvider,
		ExternalID: info.Vin,
		StyleName:  info.GetStyle(),
		SubModel:   info.GetSubModel(),
		Raw:        raw,
		FuelType:   info.FunctionResponse.Data.Decoder.FuelType.Value,
	}
	if err = validateVinDecoding(v); err != nil {
		return nil, err
	}
	return v, nil
}

func buildFromVincario(info *coremodels.VincarioInfoResponse) (*coremodels.VINDecodingInfoData, error) {
	raw, _ := json.Marshal(info)
	v := &coremodels.VINDecodingInfoData{
		VIN:        info.VIN,
		Year:       int32(info.ModelYear),
		Make:       strings.TrimSpace(info.Make),
		Model:      strings.TrimSpace(info.Model),
		Source:     coremodels.VincarioProvider,
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

	if err = validateVinDecoding(v); err != nil {
		return nil, err
	}
	return v, nil
}

func buildFromDrivly(info *coremodels.DrivlyVINResponse) (*coremodels.VINDecodingInfoData, error) {
	raw, _ := json.Marshal(info)
	yrInt, _ := strconv.Atoi(info.Year)

	v := &coremodels.VINDecodingInfoData{
		VIN:        info.Vin,
		Year:       int32(yrInt),
		Make:       info.Make,
		Model:      info.Model,
		StyleName:  buildDrivlyStyleName(info),
		ExternalID: info.GetExternalID(),
		Source:     coremodels.DrivlyProvider,
		Raw:        raw,
		FuelType:   info.Fuel,
	}
	if err := validateVinDecoding(v); err != nil {
		return nil, err
	}
	return v, nil
}

func buildDrivlyStyleName(vinInfo *coremodels.DrivlyVINResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}

// buildFromDDForTestVIN meant for use with test VIN's
func buildFromDDForTestVIN(vin string, info *coremodels.DeviceDefinitionTablelandModel) *coremodels.VINDecodingInfoData {
	makeSlug := strings.Split(info.ID, "_")[0]

	v := &coremodels.VINDecodingInfoData{
		VIN:   vin,
		Year:  int32(info.Year),
		Make:  strings.ToUpper(makeSlug[:1]) + makeSlug[1:],
		Model: info.Model,
	}

	return v
}

func buildFromDATGroup(info *coremodels.DATGroupInfoResponse) (*coremodels.VINDecodingInfoData, error) {
	if info == nil {
		return nil, fmt.Errorf("nil dat group info")
	}
	v := &coremodels.VINDecodingInfoData{
		VIN:        info.VIN,
		Year:       int32(info.Year),
		Make:       strings.TrimSpace(info.ManufacturerName),
		Model:      strings.TrimSpace(info.MainTypeGroupName),
		Source:     coremodels.DATGroupProvider,
		ExternalID: info.DatECode,
		StyleName:  info.SubModelName,
		SubModel:   info.BaseModelName,
	}
	raw, err := json.Marshal(info)
	if err == nil {
		v.Raw = raw
	}

	// todo need some more introspection from more examples
	//if strings.Contains(string(raw), "High-voltage battery") {
	//	v.FuelType = coremodels.BEV.String()
	//}

	if err := validateVinDecoding(v); err != nil {
		return nil, err
	}

	return v, nil
}

// validateVinDecoding returns an error if year, model name, make, etc seem like bad data
func validateVinDecoding(vdi *coremodels.VINDecodingInfoData) error {
	if vdi == nil {
		return fmt.Errorf("vin info was nil")
	}
	if vdi.Year == 0 || vdi.Year > int32(time.Now().Year()+1) {
		return fmt.Errorf("vin year invalid: %d", vdi.Year)
	}
	if len(vdi.Model) == 0 {
		return fmt.Errorf("vin model is empty")
	}
	if strings.Contains(vdi.Model, ",") || strings.Contains(vdi.Model, "/") {
		return fmt.Errorf("model contains invalid characters: %s", vdi.Model)
	}

	return nil
}
