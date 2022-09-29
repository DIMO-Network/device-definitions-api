//go:generate mockgen -source cache_dd.go -destination mocks/cache_dd_mock.go -package mocks

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/redis"
)

type DeviceDefinitionCacheService interface {
	GetDeviceDefinitionByID(ctx context.Context, id string) (*models.GetDeviceDefinitionQueryResult, error)
	GetDeviceDefinitionByMakeModelAndYears(ctx context.Context, make string, model string, year int) (*models.GetDeviceDefinitionQueryResult, error)
	DeleteDeviceDefinitionCacheByID(ctx context.Context, id string)
	DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx context.Context, make string, model string, year int)
}

type deviceDefinitionCacheService struct {
	Cache      redis.CacheService
	Repository repositories.DeviceDefinitionRepository
}

func NewDeviceDefinitionCacheService(cache redis.CacheService, repository repositories.DeviceDefinitionRepository) DeviceDefinitionCacheService {
	return &deviceDefinitionCacheService{Cache: cache, Repository: repository}
}

const (
	cacheLenghtHours            = 48
	cacheDeviceDefinitionKey    = "device-definition-by-id-"
	cacheDeviceDefinitionMMYKey = "device-definition-by-mmy-"
)

func (c deviceDefinitionCacheService) GetDeviceDefinitionByID(ctx context.Context, id string) (*models.GetDeviceDefinitionQueryResult, error) {

	cache := fmt.Sprintf("%s-%s", cacheDeviceDefinitionKey, id)
	cacheData := c.Cache.Get(ctx, cache)

	rp := &models.GetDeviceDefinitionQueryResult{}

	if cacheData != nil {
		val, _ := cacheData.Bytes()

		if val != nil {
			_ = json.Unmarshal(val, rp)
			return rp, nil
		}

	}

	dd, err := c.Repository.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, nil
	}

	rp = buildDeviceDefinitionResult(dd)

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLenghtHours*time.Hour)

	return rp, nil
}

func (c deviceDefinitionCacheService) DeleteDeviceDefinitionCacheByID(ctx context.Context, id string) {
	cache := fmt.Sprintf("%s-%s", cacheDeviceDefinitionKey, id)
	c.Cache.Del(ctx, cache)
}

func (c deviceDefinitionCacheService) DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx context.Context, make string, model string, year int) {
	cache := fmt.Sprintf("%s-%s-%s-%d", cacheDeviceDefinitionMMYKey, make, model, year)
	c.Cache.Del(ctx, cache)
}

func (c deviceDefinitionCacheService) GetDeviceDefinitionByMakeModelAndYears(ctx context.Context, make string, model string, year int) (*models.GetDeviceDefinitionQueryResult, error) {

	cache := fmt.Sprintf("%s-%s-%s-%d", cacheDeviceDefinitionMMYKey, make, model, year)
	cacheData := c.Cache.Get(ctx, cache)

	rp := &models.GetDeviceDefinitionQueryResult{}

	if cacheData != nil {
		val, _ := cacheData.Bytes()
		if val != nil {
			_ = json.Unmarshal(val, rp)
			return rp, nil
		}
	}

	dd, err := c.Repository.GetByMakeModelAndYears(ctx, make, model, year, true)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, nil
	}

	rp = buildDeviceDefinitionResult(dd)

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLenghtHours*time.Hour)

	return rp, nil
}

func buildDeviceDefinitionResult(dd *repoModel.DeviceDefinition) *models.GetDeviceDefinitionQueryResult {
	rp := &models.GetDeviceDefinitionQueryResult{
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

	if !dd.R.DeviceMake.TokenID.IsZero() {
		rp.DeviceMake.TokenID = dd.R.DeviceMake.TokenID.Big.Int(new(big.Int))
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
	rp.DeviceStyles = []models.GetDeviceDefinitionStylesList{}
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

		for _, ds := range dd.R.DeviceStyles {
			rp.DeviceStyles = append(rp.DeviceStyles, models.GetDeviceDefinitionStylesList{
				ID:                 ds.ID,
				DeviceDefinitionID: ds.DeviceDefinitionID,
				ExternalStyleID:    ds.ExternalStyleID,
				Name:               ds.Name,
				Source:             ds.Source,
				SubModel:           ds.SubModel,
			})
		}
	}

	return rp
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
