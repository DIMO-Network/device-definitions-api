//go:generate mockgen -source powertrain_type_service.go -destination mocks/powertrain_type_service_mock.go -package mocks

package services

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/volatiletech/null/v8"
	"os"
	"strings"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type PowerTrainTypeService interface {
	ResolvePowerTrainType(ctx context.Context, makeSlug string, modelSlug string, definitionID *string, drivlyData null.JSON, vincarioData null.JSON) (*string, error)
}

type powerTrainTypeService struct {
	DBS                func() *db.ReaderWriter
	logger             *zerolog.Logger
	powerTrainRuleData coremodels.PowerTrainTypeRuleData
}

func NewPowerTrainTypeService(dbs func() *db.ReaderWriter, logger *zerolog.Logger) (PowerTrainTypeService, error) {
	content, err := os.ReadFile("powertrain_type_rule.yaml")
	if err != nil {
		return nil, err
	}

	var powerTrainTypeData coremodels.PowerTrainTypeRuleData
	err = yaml.Unmarshal(content, &powerTrainTypeData)
	if err != nil {
		return nil, err
	}

	return &powerTrainTypeService{DBS: dbs, logger: logger, powerTrainRuleData: powerTrainTypeData}, nil
}

// todo needs a test

func (c powerTrainTypeService) ResolvePowerTrainType(ctx context.Context, makeSlug string, modelSlug string, definitionID *string, drivlyData null.JSON, vincarioData null.JSON) (*string, error) {

	for _, ptType := range c.powerTrainRuleData.PowerTrainTypeList {
		for _, mk := range ptType.Makes {
			if mk == makeSlug {
				if len(ptType.Models) == 0 {
					return &ptType.Type, nil
				}

				for _, model := range ptType.Models {
					if model == modelSlug {
						return &ptType.Type, nil
					}
				}

			}
		}
	}
	// model name based inference
	if strings.Contains(modelSlug, "plug-in") {
		p := coremodels.PHEV.String()
		return &p, nil
	}
	if strings.Contains(modelSlug, "hybrid") {
		p := coremodels.HEV.String()
		return &p, nil
	}

	// Default
	defaultPowerTrainType := ""
	for _, ptType := range c.powerTrainRuleData.PowerTrainTypeList {
		if ptType.Default {
			defaultPowerTrainType = ptType.Type
			break
		}
	}
	// if definitionId is not nil set the drivlyData & vincarioData from a vin number that matches ddID
	if definitionID != nil {
		vins, err := models.VinNumbers(models.VinNumberWhere.DeviceDefinitionID.EQ(*definitionID)).All(ctx, c.DBS().Reader)
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
			c.logger.Error().Err(err)
		}
		c.logger.Info().Msg("Looking up PowerTrain from Drivly Data")
		for _, item := range c.powerTrainRuleData.DrivlyList {
			if len(item.Values) > 0 {
				for _, value := range item.Values {
					if value == drivlyModel.FuelType {
						return &item.Type, nil
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
			c.logger.Error().Err(err)
		}
		c.logger.Info().Msg("Looking up PowerTrain from Vincario Data")
		for _, item := range c.powerTrainRuleData.VincarioList {
			if len(item.Values) > 0 {
				for _, value := range item.Values {
					if value == vincarioModel.FuelType {
						return &item.Type, nil
					}
				}
			}
		}
	}

	return &defaultPowerTrainType, nil
}
