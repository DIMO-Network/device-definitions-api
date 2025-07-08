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

	"github.com/DIMO-Network/shared/pkg/logfields"

	"github.com/DIMO-Network/shared/pkg/db"

	vinutil "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/aarondl/null/v8"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type VINDecodingService interface {
	// GetVIN decodes a vin using one of the providers passed in or if AllProviders applies an ordered logic. Only pass TeslaProvider if know it is a Tesla.
	GetVIN(ctx context.Context, vin string, provider coremodels.DecodeProviderEnum, country string) (*coremodels.VINDecodingInfoData, error)
}

type vinDecodingService struct {
	logger             *zerolog.Logger
	drivlyAPISvc       gateways.DrivlyAPIService
	vincarioAPISvc     gateways.VincarioAPIService
	autoIsoAPIService  gateways.AutoIsoAPIService
	DATGroupAPIService gateways.DATGroupAPIService
	japan17VINAPI      gateways.Japan17VINAPI
	carvxAPI           gateways.CarVxVINAPI
	onChainSvc         gateways.DeviceDefinitionOnChainService
	dbs                func() *db.ReaderWriter
}

func NewVINDecodingService(drivlyAPISvc gateways.DrivlyAPIService, vincarioAPISvc gateways.VincarioAPIService, autoIso gateways.AutoIsoAPIService, logger *zerolog.Logger,
	onChainSvc gateways.DeviceDefinitionOnChainService, datGroupAPIService gateways.DATGroupAPIService, dbs func() *db.ReaderWriter,
	japan17VINAPI gateways.Japan17VINAPI, carvxAPI gateways.CarVxVINAPI) VINDecodingService {
	return &vinDecodingService{drivlyAPISvc: drivlyAPISvc, vincarioAPISvc: vincarioAPISvc, autoIsoAPIService: autoIso,
		japan17VINAPI: japan17VINAPI, carvxAPI: carvxAPI, logger: logger, onChainSvc: onChainSvc, DATGroupAPIService: datGroupAPIService, dbs: dbs}
}

func (c vinDecodingService) GetVIN(ctx context.Context, vin string, provider coremodels.DecodeProviderEnum, country string) (*coremodels.VINDecodingInfoData, error) {
	const DefaultDefinitionID = "ford_escape_2020"

	result := &coremodels.VINDecodingInfoData{}
	vin = strings.ToUpper(strings.TrimSpace(vin))
	providersToTry := make([]coremodels.DecodeProviderEnum, 0)
	// check for japan chasis if all providers
	if provider == coremodels.AllProviders && ((len(vin) < 17 && len(vin) > 10) || country == "JPN") {
		providersToTry = append(providersToTry, coremodels.CarVXVIN)
		providersToTry = append(providersToTry, coremodels.Japan17VIN)
	} else if !ValidateVIN(vin) {
		return nil, fmt.Errorf("invalid vin: %s", vin)
	}

	localLog := c.logger.With().
		Str(logfields.VIN, vin).
		Str(logfields.FunctionName, "GetVIN").
		Logger()

	if strings.HasPrefix(vin, "0SC") {
		dd, _, err := c.onChainSvc.GetDefinitionByID(ctx, DefaultDefinitionID)
		if err != nil {
			return nil, err
		}
		result = buildFromDDForTestVIN(vin, dd)
		return result, nil
	}

	if len(providersToTry) == 0 {
		if provider == coremodels.AllProviders {
			// fill in the list, future could do something country specific
			providersToTry = append(providersToTry, coremodels.DrivlyProvider, coremodels.VincarioProvider, coremodels.Japan17VIN, coremodels.AutoIsoProvider, coremodels.DATGroupProvider)
		} else {
			// use the specified override
			providersToTry = append(providersToTry, provider)
		}
	}
	var errFinal error // for later

	for _, p := range providersToTry {
		// try all the options, but need to continue if get an error
		switch p {
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
			err := validateVinDecoding(result)
			if err != nil {
				errFinal = err
				continue
			}
			localLog.Info().Msgf("decoded with tesla: %+v", result)
			return result, nil
		case coremodels.DrivlyProvider:
			localLog.Info().Msgf("trying to decode VIN: %s with drivly", vin)
			vinDrivlyInfo, err := c.drivlyAPISvc.GetVINInfo(vin)
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with drivly", vin)
				continue
			}
			result, err = buildFromDrivly(vinDrivlyInfo) // already does validation
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with drivly", vin)
				continue
			}
			return result, nil
		case coremodels.VincarioProvider:
			localLog.Info().Msgf("trying to decode VIN: %s with vincario", vin)
			vinVincarioInfo, err := c.vincarioAPISvc.DecodeVIN(vin)
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with vincario", vin)
				continue
			}
			result, err = buildFromVincario(vinVincarioInfo) // already does validation
			if err != nil {
				errFinal = err
				continue
			}
			return result, nil
		case coremodels.Japan17VIN:
			localLog.Info().Msgf("trying to decode VIN: %s with 17vin", vin)
			mmy, raw, err := c.japan17VINAPI.GetVINInfo(vin)
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with japan17vin", vin)
				continue
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
			err = validateVinDecoding(result)
			if err != nil {
				errFinal = err
				continue
			}
			return result, nil
		case coremodels.CarVXVIN:
			localLog.Info().Msgf("trying to decode VIN: %s with carvx", vin)
			info, raw, err := c.carvxAPI.GetVINInfo(vin)
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with carvx", vin)
				continue
			}
			yr, _ := strconv.Atoi(info.Data[0].ManufactureDate.Year)
			result = &coremodels.VINDecodingInfoData{
				VIN:       vin,
				Make:      info.Data[0].Make,
				Model:     info.Data[0].Model,
				SubModel:  info.Data[0].Drive + " " + info.Data[0].Transmission,
				Year:      int32(yr),
				StyleName: info.Data[0].Drive + " " + info.Data[0].Transmission,
				Source:    coremodels.CarVXVIN,
				MetaData:  null.JSONFrom(raw),
				Raw:       raw,
				FuelType:  info.Data[0].Fuel,
			}
			err = validateVinDecoding(result)
			if err != nil {
				errFinal = err
				continue
			}
			return result, nil
		case coremodels.AutoIsoProvider:
			localLog.Info().Msgf("trying to decode VIN: %s with autoiso", vin)
			vinAutoIsoInfo, err := c.autoIsoAPIService.GetVIN(vin)
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with autoiso", vin)
				continue
			}
			result, err = buildFromAutoIso(vinAutoIsoInfo) // already does validation
			if err != nil {
				errFinal = err
				continue
			}
			return result, nil
		case coremodels.DATGroupProvider:
			localLog.Info().Msgf("trying to decode VIN: %s with datgroup", vin)
			// todo lookup country for two letter equiv
			vinInfo, err := c.DATGroupAPIService.GetVINv2(vin, country) // try with Turkey
			if err != nil {
				errFinal = errors.Wrapf(err, "unable to decode vin: %s with DATGroup", vin)
				continue
			}
			result, err = buildFromDATGroup(vinInfo) // already does validation
			if err != nil {
				errFinal = err
				continue
			}
			return result, nil
		case coremodels.AllProviders:
			// this should never hit
			errFinal = fmt.Errorf("all providers - invalid option reached")
		}
	}

	// could not decode anything
	if errFinal != nil {
		return nil, errFinal
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
	if vdi.Source == "" {
		return fmt.Errorf("vin source is empty")
	}

	return nil
}
