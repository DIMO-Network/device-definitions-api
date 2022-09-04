package services

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"

	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
)

type DeviceDefinitionCacheService interface {
	GetDeviceDefinitionByID(ctx context.Context, id string) (*models.GetDeviceDefinitionQueryResult, error)
	GetDeviceDefinitionByMakeModelAndYears(ctx context.Context, make string, model string, year int) (*models.GetDeviceDefinitionQueryResult, error)
}

type deviceDefinitionCacheService struct {
	Cache      gateways.RedisCacheService
	Repository repositories.DeviceDefinitionRepository
}

func NewDeviceDefinitionCacheService(cache gateways.RedisCacheService, repository repositories.DeviceDefinitionRepository) DeviceDefinitionCacheService {
	return &deviceDefinitionCacheService{Cache: cache, Repository: repository}
}

func (c deviceDefinitionCacheService) GetDeviceDefinitionByID(ctx context.Context, id string) (*models.GetDeviceDefinitionQueryResult, error) {

	dd, _ := c.Repository.GetByID(ctx, id)

	if dd == nil {
		return nil, fmt.Errorf("could not find device definition id: %s", id)
	}

	rp := models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.String,
		CompatibleIntegrations: []models.GetDeviceCompatibility{},
		DeviceMake: models.DeviceMake{
			ID:              dd.R.DeviceMake.ID,
			Name:            dd.R.DeviceMake.Name,
			LogoURL:         dd.R.DeviceMake.LogoURL,
			OemPlatformName: dd.R.DeviceMake.OemPlatformName,
		},
		Type: models.DeviceType{
			Type:  "Vehicle",
			Make:  dd.R.DeviceMake.Name,
			Model: dd.Model,
			Year:  int(dd.Year),
		},
		Metadata: string(dd.Metadata.JSON),
		Verified: dd.Verified,
	}

	// vehicle info
	var vi map[string]models.GetDeviceVehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		rp.VehicleInfo = vi["vehicle_info"]
	}

	if dd.R != nil {
		// compatible integrations
		rp.CompatibleIntegrations = deviceCompatibilityFromDB(dd.R.DeviceIntegrations)
		// sub_models
		rp.Type.SubModels = common.SubModelsFromStylesDB(dd.R.DeviceStyles)
	}

	// build object for integrations that have all the info
	rp.DeviceIntegrations = []models.GetDeviceDefinitionIntegrationList{}
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.DeviceIntegrations = append(rp.DeviceIntegrations, models.GetDeviceDefinitionIntegrationList{
				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendor,
				Region:       di.Region,
				Capabilities: common.JSONOrDefault(di.Capabilities),
			})
		}
	}

	return &models.GetDeviceDefinitionQueryResult{}, nil
}

func (c deviceDefinitionCacheService) GetDeviceDefinitionByMakeModelAndYears(ctx context.Context, make string, model string, year int) (*models.GetDeviceDefinitionQueryResult, error) {

	dd, _ := c.Repository.GetByMakeModelAndYears(ctx, make, model, year, true)

	if dd == nil {
		return nil, nil
	}

	rp := models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.String,
		CompatibleIntegrations: []models.GetDeviceCompatibility{},
		DeviceMake: models.DeviceMake{
			ID:              dd.R.DeviceMake.ID,
			Name:            dd.R.DeviceMake.Name,
			LogoURL:         dd.R.DeviceMake.LogoURL,
			OemPlatformName: dd.R.DeviceMake.OemPlatformName,
		},
		Type: models.DeviceType{
			Type:  "Vehicle",
			Make:  dd.R.DeviceMake.Name,
			Model: dd.Model,
			Year:  int(dd.Year),
		},
		Metadata: string(dd.Metadata.JSON),
		Verified: dd.Verified,
	}

	// vehicle info
	var vi map[string]models.GetDeviceVehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		rp.VehicleInfo = vi["vehicle_info"]
	}

	if dd.R != nil {
		// compatible integrations
		rp.CompatibleIntegrations = deviceCompatibilityFromDB(dd.R.DeviceIntegrations)
		// sub_models
		rp.Type.SubModels = common.SubModelsFromStylesDB(dd.R.DeviceStyles)
	}

	// build object for integrations that have all the info
	rp.DeviceIntegrations = []models.GetDeviceDefinitionIntegrationList{}
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.DeviceIntegrations = append(rp.DeviceIntegrations, models.GetDeviceDefinitionIntegrationList{
				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendor,
				Region:       di.Region,
				Capabilities: common.JSONOrDefault(di.Capabilities),
			})
		}
	}

	return &models.GetDeviceDefinitionQueryResult{}, nil
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func deviceCompatibilityFromDB(dbDIS repoModel.DeviceIntegrationSlice) []models.GetDeviceCompatibility {
	if len(dbDIS) == 0 {
		return []models.GetDeviceCompatibility{}
	}
	compatibilities := make([]models.GetDeviceCompatibility, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = models.GetDeviceCompatibility{
			ID:           di.IntegrationID,
			Type:         di.R.Integration.Type,
			Style:        di.R.Integration.Style,
			Vendor:       di.R.Integration.Vendor,
			Region:       di.Region,
			Capabilities: common.JSONOrDefault(di.Capabilities),
		}
	}
	return compatibilities
}
