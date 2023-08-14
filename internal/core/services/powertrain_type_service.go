//go:generate mockgen -source powertrain_type_service.go -destination mocks/powertrain_type_service_mock.go -package mocks

package services

import (
	"context"
	"os"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type PowerTrainTypeService interface {
	ResolvePowerTrainType(ctx context.Context, makeSlug string, modelSlug string, definition *models.DeviceDefinition) (*string, error)
}

type powerTrainTypeService struct {
	DBS    func() *db.ReaderWriter
	logger *zerolog.Logger
}

func NewPowerTrainTypeService(dbs func() *db.ReaderWriter, logger *zerolog.Logger) PowerTrainTypeService {
	return &powerTrainTypeService{DBS: dbs, logger: logger}
}

func (c powerTrainTypeService) ResolvePowerTrainType(ctx context.Context, makeSlug string, modelSlug string, definition *models.DeviceDefinition) (*string, error) {

	content, err := os.ReadFile("powertrain_type_rule.yaml")
	if err != nil {
		return nil, err
	}

	var powerTrainTypeData coremodels.PowerTrainTypeRuleData
	err = yaml.Unmarshal(content, &powerTrainTypeData)
	if err != nil {
		return nil, err
	}

	for _, ptType := range powerTrainTypeData.PowerTrainTypeList {
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

	// Default
	defaultPowerTrainType := ""
	for _, ptType := range powerTrainTypeData.PowerTrainTypeList {
		if ptType.Default {
			defaultPowerTrainType = ptType.Type
			break
		}
	}

	if definition != nil {
		vins, err := models.VinNumbers(models.VinNumberWhere.DeviceDefinitionID.EQ(definition.ID)).All(ctx, c.DBS().Reader)
		if err != nil {
			return &defaultPowerTrainType, nil
		}

		if len(vins) == 0 {
			return &defaultPowerTrainType, nil
		}

		vin := vins[0]

		// Resolve Drivly Data
		c.logger.Info().Msg("Looking up PowerTrain from Drivly Data")
		if vin.DrivlyData.Valid && len(powerTrainTypeData.VincarioList) > 0 {
			var drivlyData coremodels.DrivlyData
			err = vin.DrivlyData.Unmarshal(&drivlyData)
			if err != nil {
				c.logger.Error().Err(err)
			}

			for _, item := range powerTrainTypeData.DrivlyList {
				if len(item.Values) > 0 {
					for _, value := range item.Values {
						if value == drivlyData.FuelType {
							return &item.Type, nil
						}
					}
				}
			}

		}

		// Resolve Vincario Data
		c.logger.Info().Msg("Looking up PowerTrain from Vincario Data")
		if vin.VincarioData.Valid && len(powerTrainTypeData.VincarioList) > 0 {
			var vincarioData coremodels.VincarioData
			err = vin.DrivlyData.Unmarshal(&vincarioData)
			if err != nil {
				c.logger.Error().Err(err)
			}

			for _, item := range powerTrainTypeData.VincarioList {
				if len(item.Values) > 0 {
					for _, value := range item.Values {
						if value == vincarioData.FuelType {
							return &item.Type, nil
						}
					}
				}
			}
		}
	}

	return &defaultPowerTrainType, nil
}
