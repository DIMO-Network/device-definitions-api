//go:generate mockgen -source powertrain_type_service.go -destination mocks/powertrain_type_service_mock.go -package mocks

package services

import (
	"context"
	"os"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/volatiletech/null/v8"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type PowerTrainTypeService interface {
	ResolvePowerTrainType(ctx context.Context, makeSlug string, modelSlug string, definitionID *string, drivlyData null.JSON, vincarioData null.JSON) (string, error)
	ResolvePowerTrainFromVinInfo(vinInfo *coremodels.VINDecodingInfoData) string
}

type powerTrainTypeService struct {
	DBS                            func() *db.ReaderWriter
	logger                         *zerolog.Logger
	powerTrainRuleData             coremodels.PowerTrainTypeRuleData
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService
}

func NewPowerTrainTypeService(dbs func() *db.ReaderWriter, rulesFileName string, logger *zerolog.Logger, ddOnChainSvc gateways.DeviceDefinitionOnChainService) (PowerTrainTypeService, error) {
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

	return &powerTrainTypeService{DBS: dbs, logger: logger, powerTrainRuleData: powerTrainTypeData, deviceDefinitionOnChainService: ddOnChainSvc}, nil
}

// ResolvePowerTrainFromVinInfo uses standard vin info StyleName and FuelType to figure out powertrain, otherwise returns an empty string
func (c powerTrainTypeService) ResolvePowerTrainFromVinInfo(vinInfo *coremodels.VINDecodingInfoData) string {
	// style name based inference
	pt := powertrainNameInference(vinInfo.StyleName)
	if pt != "" {
		return pt
	}
	// we may need a parameter for the provider type and then case below
	// drivly loop, using fuel type to try to get powertrain
	for _, item := range c.powerTrainRuleData.DrivlyList {
		if len(item.Values) > 0 {
			for _, value := range item.Values {
				if value == vinInfo.FuelType {
					return item.Type
				}
			}
		}
	}
	// loop over for vincario
	for _, item := range c.powerTrainRuleData.VincarioList {
		if len(item.Values) > 0 {
			for _, value := range item.Values {
				if value == vinInfo.FuelType {
					return item.Type
				}
			}
		}
	}
	// return empty if no matches
	return ""
}

// ResolvePowerTrainType figures out the powertrain based on make, model, and optionally definitionID or vin decoder data
func (c powerTrainTypeService) ResolvePowerTrainType(ctx context.Context, makeSlug string, modelSlug string, definitionID *string, drivlyData null.JSON, vincarioData null.JSON) (string, error) {
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

	if definitionID != nil {
		// future: what about style. also, use the dd cache service
		// first see if we have already figured out powertrain for this DD
		dd, errTbl := c.deviceDefinitionOnChainService.GetDefinitionByID(ctx, *definitionID, c.DBS().Reader)
		if dd != nil {
			c.logger.Warn().Err(errTbl).Msgf("failed to get dd from tableland node: %s", *definitionID)
		}
		if dd != nil && dd.Metadata != nil {
			for _, attr := range dd.Metadata.DeviceAttributes {
				if attr.Name == common.PowerTrainType {
					return attr.Value, nil
				}
			}
		}
		// if definitionId is not nil set the drivlyData & vincarioData from a vin number that matches ddID
		vins, err := models.VinNumbers(models.VinNumberWhere.DefinitionID.EQ(*definitionID)).All(ctx, c.DBS().Reader)
		if err == nil && len(vins) > 0 {
			drivlyData = vins[0].DrivlyData
			vincarioData = vins[0].VincarioData
		}
	}

	// Resolve Drivly Data
	if drivlyData.Valid && len(c.powerTrainRuleData.VincarioList) > 0 {
		var drivlyModel coremodels.DrivlyData
		err := drivlyData.Unmarshal(&drivlyModel)
		if err != nil {
			c.logger.Error().Err(err).Send()
		}
		c.logger.Info().Msg("Looking up PowerTrain from Drivly Data")
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
		c.logger.Info().Msg("Looking up PowerTrain from Vincario Data")
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

// powertrainNameInference figures out powertrain just from name
func powertrainNameInference(name string) string {
	// model name based inference
	if strings.Contains(name, "plug-in") {
		return coremodels.PHEV.String()
	}
	if strings.Contains(name, "hybrid") {
		return coremodels.HEV.String()
	}
	if strings.Contains(name, "e-tron") {
		return coremodels.BEV.String()
	}

	return ""
}
