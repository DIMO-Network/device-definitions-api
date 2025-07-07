//go:generate mockgen -source powertrain_type_service.go -destination mocks/powertrain_type_service_mock.go -package mocks

package services

import (
	"os"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/volatiletech/null/v8"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type PowerTrainTypeService interface {
	ResolvePowerTrainType(makeSlug string, modelSlug string, drivlyData null.JSON, vincarioData null.JSON) (string, error)
	ResolvePowerTrainFromVinInfo(styleName, fuelType string) string
}

type powerTrainTypeService struct {
	logger                         *zerolog.Logger
	powerTrainRuleData             coremodels.PowerTrainTypeRuleData
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService
}

func NewPowerTrainTypeService(rulesFileName string, logger *zerolog.Logger, ddOnChainSvc gateways.DeviceDefinitionOnChainService) (PowerTrainTypeService, error) {
	if rulesFileName == "" {
		rulesFileName = "powertrain_type_rule.yaml"
	}
	content, err := os.ReadFile(rulesFileName)
	if err != nil {
		return nil, err
	}

	var powerTrainTypeData coremodels.PowerTrainTypeRuleData
	err = yaml.Unmarshal(content, &powerTrainTypeData)
	if err != nil {
		return nil, err
	}

	return &powerTrainTypeService{logger: logger, powerTrainRuleData: powerTrainTypeData, deviceDefinitionOnChainService: ddOnChainSvc}, nil
}

// ResolvePowerTrainFromVinInfo uses standard vin info StyleName and FuelType to figure out powertrain, otherwise returns an empty string
func (c powerTrainTypeService) ResolvePowerTrainFromVinInfo(styleName, fuelType string) string {
	// style name based inference
	pt := powertrainNameInference(styleName)
	if pt != "" {
		return pt
	}
	if fuelType == "" {
		return ""
	}
	// we may need a parameter for the provider type and then case below
	// drivly loop, using fuel type to try to get powertrain
	for _, item := range c.powerTrainRuleData.DrivlyList {
		if len(item.Values) > 0 {
			for _, value := range item.Values {
				if value == fuelType {
					return item.Type
				}
			}
		}
	}
	// loop over for vincario
	for _, item := range c.powerTrainRuleData.VincarioList {
		if len(item.Values) > 0 {
			for _, value := range item.Values {
				if value == fuelType {
					return item.Type
				}
			}
		}
	}
	// return empty if no matches
	return ""
}

// ResolvePowerTrainType figures out the powertrain based on make, model, and optionally definitionID or vin decoder data
func (c powerTrainTypeService) ResolvePowerTrainType(makeSlug string, modelSlug string, drivlyData null.JSON, vincarioData null.JSON) (string, error) {
	if makeSlug == "tesla" || makeSlug == "rivian" || makeSlug == "lucid" {
		return coremodels.BEV.String(), nil
	}
	// iterate over hard coded rules in yaml file first
	for _, ptType := range c.powerTrainRuleData.PowerTrainTypeList {
		for _, mk := range ptType.Makes {
			if mk == makeSlug {
				if len(ptType.Models) == 0 {
					return ptType.Type, nil
				}

				for _, model := range ptType.Models {
					if model == modelSlug {
						return ptType.Type, nil
					}
				}
			}
		}
	}
	pt := powertrainNameInference(modelSlug)
	if pt != "" {
		return pt, nil
	}

	// Default
	defaultPowerTrainType := ""
	for _, ptType := range c.powerTrainRuleData.PowerTrainTypeList {
		if ptType.Default {
			defaultPowerTrainType = ptType.Type
			break
		}
	}

	// Resolve Drivly Data
	if drivlyData.Valid && len(c.powerTrainRuleData.DrivlyList) > 0 {
		var drivlyModel coremodels.DrivlyData
		err := drivlyData.Unmarshal(&drivlyModel)
		if err != nil {
			c.logger.Error().Err(err).Send()
		}
		c.logger.Debug().Msgf("Looking up PowerTrain from Drivly Data for %s", modelSlug)
		for _, item := range c.powerTrainRuleData.DrivlyList {
			if len(item.Values) > 0 {
				for _, value := range item.Values {
					if value == drivlyModel.Fuel {
						return item.Type, nil
					}
				}
			}
		}
	}
	// Resolve Vincario Data
	if vincarioData.Valid && len(c.powerTrainRuleData.VincarioList) > 0 {
		var vincarioModel coremodels.VincarioData
		err := vincarioData.Unmarshal(&vincarioModel)
		if err != nil {
			c.logger.Error().Err(err).Send()
		}
		c.logger.Debug().Msgf("Looking up PowerTrain from Vincario Data for %s", modelSlug)
		for _, item := range c.powerTrainRuleData.VincarioList {
			if len(item.Values) > 0 {
				for _, value := range item.Values {
					if value == vincarioModel.FuelType {
						return item.Type, nil
					}
				}
			}
		}
	}

	if defaultPowerTrainType == "" {
		defaultPowerTrainType = coremodels.ICE.String()
	}

	return defaultPowerTrainType, nil
}

// powertrainNameInference figures out powertrain just from name based on common patterns
func powertrainNameInference(modelSlug string) string {
	// model modelSlug based inference
	if strings.Contains(modelSlug, "plug-in") {
		return coremodels.PHEV.String()
	}
	if strings.Contains(modelSlug, "hybrid") {
		return coremodels.HEV.String()
	}
	if strings.Contains(modelSlug, "e-tron") {
		return coremodels.BEV.String()
	}
	if strings.Contains(modelSlug, "-ev") {
		return coremodels.BEV.String()
	}

	return ""
}
