//go:generate mockgen -source cache_dd.go -destination mocks/cache_dd_mock.go -package mocks

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/shared"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"sync"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	repomodels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/redis"
)

type DeviceDefinitionCacheService interface {
	GetDeviceMakeByName(ctx context.Context, makeName string) (*models.DeviceMake, error)
	GetDeviceDefinitionByID(ctx context.Context, id string, options ...GetDeviceDefinitionOption) (*models.GetDeviceDefinitionQueryResult, error)
	GetDeviceDefinitionBySlug(ctx context.Context, slug string, year int) (*models.GetDeviceDefinitionQueryResult, error)
	GetDeviceDefinitionByMakeModelAndYears(ctx context.Context, mk string, model string, year int) (*models.GetDeviceDefinitionQueryResult, error)
	DeleteDeviceDefinitionCacheByID(ctx context.Context, id string)
	DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx context.Context, mk string, model string, year int)
	DeleteDeviceDefinitionCacheBySlug(ctx context.Context, slug string, year int)
	GetDeviceDefinitionBySlugName(ctx context.Context, slug string) (*models.GetDeviceDefinitionQueryResult, error)
}

type deviceDefinitionCacheService struct {
	Cache         redis.CacheService
	Repository    repositories.DeviceDefinitionRepository
	makesRepo     repositories.DeviceMakeRepository
	ddOnChainSvc  gateways.DeviceDefinitionOnChainService
	queryInstance *contracts.Registry
	makesBySlug   map[string]*models.DeviceMake
	mu            sync.RWMutex
}

func NewDeviceDefinitionCacheService(cache redis.CacheService, repository repositories.DeviceDefinitionRepository, deviceMakesRepo repositories.DeviceMakeRepository,
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService, registryInstance *contracts.Registry) DeviceDefinitionCacheService {
	return &deviceDefinitionCacheService{Cache: cache, Repository: repository, ddOnChainSvc: deviceDefinitionOnChainService,
		queryInstance: registryInstance, makesRepo: deviceMakesRepo, makesBySlug: make(map[string]*models.DeviceMake)}
}

const (
	cacheLengthHours                 = 48
	cacheDeviceDefinitionKey         = "device-definition-by-id-"
	cacheDeviceDefinitionMMYKey      = "device-definition-by-mmy-"
	cacheDeviceDefinitionSlugKey     = "device-definition-by-slug-"
	cacheDeviceDefinitionSlugNameKey = "device-definition-by-slug-name"
)

func (c *deviceDefinitionCacheService) GetDeviceMakeByName(ctx context.Context, makeName string) (*models.DeviceMake, error) {
	slug := shared.SlugString(makeName)

	if m, ok := c.makesBySlug[slug]; ok {
		return m, nil
	}
	// else need to get and set
	c.mu.Lock()
	defer c.mu.Unlock()

	dm, err := c.makesRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	manufacturerID, err := c.queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, makeName)
	if err != nil {
		return nil, err
	}
	result := &models.DeviceMake{
		ID:              dm.ID,
		Name:            dm.Name,
		LogoURL:         dm.LogoURL,
		OemPlatformName: dm.OemPlatformName,
		TokenID:         manufacturerID,
		NameSlug:        dm.NameSlug,
		Metadata:        dm.Metadata.JSON,
		MetadataTyped:   common.BuildDeviceMakeMetadata(dm.Metadata),
		CreatedAt:       dm.CreatedAt,
		UpdatedAt:       dm.UpdatedAt,
	}
	c.makesBySlug[slug] = result

	return result, nil
}

func (c *deviceDefinitionCacheService) GetDeviceDefinitionByID(ctx context.Context, id string, opts ...GetDeviceDefinitionOption) (*models.GetDeviceDefinitionQueryResult, error) {

	params := defaultGetDeviceDefinitionCacheOptions
	for _, opt := range opts {
		opt(&params)
	}

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

	var dd *repomodels.DeviceDefinition
	var err error

	if !params.UseOnChainData {
		dd, err = c.Repository.GetByID(ctx, id)

		if err != nil {
			return nil, err
		}

		if dd == nil {
			return nil, nil
		}

		rp, err = common.BuildFromDeviceDefinitionToQueryResult(dd)
		if err != nil {
			return nil, err
		}
	}

	if params.UseOnChainData {
		dm, err := c.GetDeviceMakeByName(ctx, params.Make.Name)
		if err != nil {
			return nil, err
		}
		dd, err = c.ddOnChainSvc.GetDeviceDefinitionByID(ctx, dm.TokenID, id)

		if err != nil {
			return nil, err
		}

		if dd == nil {
			return nil, nil
		}

		dd.R = dd.R.NewStruct()
		dd.R.DeviceMake = params.Make
		dd.R.DeviceType = &repomodels.DeviceType{
			Metadatakey: common.VehicleMetadataKey,
		}
		rp, err = common.BuildFromDeviceDefinitionToQueryResult(dd)
		if err != nil {
			return nil, err
		}

	}

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLengthHours*time.Hour)

	return rp, nil
}

func (c *deviceDefinitionCacheService) DeleteDeviceDefinitionCacheByID(ctx context.Context, id string) {
	cache := fmt.Sprintf("%s-%s", cacheDeviceDefinitionKey, id)
	c.Cache.Del(ctx, cache)
}

func (c *deviceDefinitionCacheService) DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx context.Context, mk string, model string, year int) {
	cache := fmt.Sprintf("%s-%s-%s-%d", cacheDeviceDefinitionMMYKey, mk, model, year)
	c.Cache.Del(ctx, cache)
}

func (c *deviceDefinitionCacheService) GetDeviceDefinitionByMakeModelAndYears(ctx context.Context, mk string, model string, year int) (*models.GetDeviceDefinitionQueryResult, error) {

	cache := fmt.Sprintf("%s-%s-%s-%d", cacheDeviceDefinitionMMYKey, mk, model, year)
	cacheData := c.Cache.Get(ctx, cache)

	rp := &models.GetDeviceDefinitionQueryResult{}

	if cacheData != nil {
		val, _ := cacheData.Bytes()
		if val != nil {
			_ = json.Unmarshal(val, rp)
			return rp, nil
		}
	}

	dd, err := c.Repository.GetByMakeModelAndYears(ctx, mk, model, year, true)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, nil
	}

	rp, err = common.BuildFromDeviceDefinitionToQueryResult(dd)
	if err != nil {
		return nil, err
	}

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLengthHours*time.Hour)

	return rp, nil
}

func (c *deviceDefinitionCacheService) GetDeviceDefinitionBySlug(ctx context.Context, slug string, year int) (*models.GetDeviceDefinitionQueryResult, error) {

	cache := fmt.Sprintf("%s-%s-%d", cacheDeviceDefinitionSlugKey, slug, year)
	cacheData := c.Cache.Get(ctx, cache)

	rp := &models.GetDeviceDefinitionQueryResult{}

	if cacheData != nil {
		val, _ := cacheData.Bytes()
		if val != nil {
			_ = json.Unmarshal(val, rp)
			return rp, nil
		}
	}

	dd, err := c.Repository.GetBySlugAndYear(ctx, slug, year, true)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, nil
	}

	rp, err = common.BuildFromDeviceDefinitionToQueryResult(dd)
	if err != nil {
		return nil, err
	}

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLengthHours*time.Hour)

	return rp, nil
}

func (c *deviceDefinitionCacheService) GetDeviceDefinitionBySlugName(ctx context.Context, slug string) (*models.GetDeviceDefinitionQueryResult, error) {

	cache := fmt.Sprintf("%s-%s", cacheDeviceDefinitionSlugNameKey, slug)
	cacheData := c.Cache.Get(ctx, cache)

	rp := &models.GetDeviceDefinitionQueryResult{}

	if cacheData != nil {
		val, _ := cacheData.Bytes()
		if val != nil {
			_ = json.Unmarshal(val, rp)
			return rp, nil
		}
	}

	dd, err := c.Repository.GetBySlugName(ctx, slug, true)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, nil
	}

	rp, err = common.BuildFromDeviceDefinitionToQueryResult(dd)
	if err != nil {
		return nil, err
	}

	rpJSON, _ := json.Marshal(rp)
	_ = c.Cache.Set(ctx, cache, rpJSON, cacheLengthHours*time.Hour)

	return rp, nil
}

func (c *deviceDefinitionCacheService) DeleteDeviceDefinitionCacheBySlug(ctx context.Context, slug string, year int) {
	cache := fmt.Sprintf("%s-%s-%d", cacheDeviceDefinitionSlugKey, slug, year)
	c.Cache.Del(ctx, cache)
}

type GetDeviceDefinitionCacheOptions struct {
	UseOnChainData bool
	Make           *repomodels.DeviceMake
}

var defaultGetDeviceDefinitionCacheOptions = GetDeviceDefinitionCacheOptions{
	UseOnChainData: false,
}

type GetDeviceDefinitionOption func(*GetDeviceDefinitionCacheOptions)

func UseOnChain(deviceMake *repomodels.DeviceMake) GetDeviceDefinitionOption {
	return func(opts *GetDeviceDefinitionCacheOptions) {
		opts.UseOnChainData = true
		opts.Make = deviceMake
	}
}
