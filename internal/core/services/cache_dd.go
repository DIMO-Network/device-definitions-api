//go:generate mockgen -source cache_dd.go -destination mocks/cache_dd_mock.go -package mocks

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
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

	rp = common.BuildFromDeviceDefinitionToQueryResult(dd)

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

	rp = common.BuildFromDeviceDefinitionToQueryResult(dd)

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLenghtHours*time.Hour)

	return rp, nil
}
