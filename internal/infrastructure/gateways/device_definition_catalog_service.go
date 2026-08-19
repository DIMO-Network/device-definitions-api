//nolint:tagliatelle
package gateways

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source device_definition_catalog_service.go -destination mocks/device_definition_catalog_service_mock.go -package mocks

// DeviceDefinitionCatalogService reads device definitions from the R2-backed
// catalog (CDN) and writes them through the definitions-worker. It replaces
// the Tableland on-chain service.
type DeviceDefinitionCatalogService interface {
	GetManufacturer(manufacturerSlug string) (*coremodels.Manufacturer, error)
	GetManufacturerNameByID(ctx context.Context, manufacturerID *big.Int) (string, error)
	// GetDeviceDefinitionByID gets a definition by slug ID, requiring it to belong to the given manufacturer.
	GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	// GetDefinitionByID gets a definition by slug ID and returns the manufacturer token id too.
	GetDefinitionByID(ctx context.Context, ID string) (*coremodels.DeviceDefinitionTablelandModel, *big.Int, error)
	// GetDefinition is GetDeviceDefinitionByID under its historical secondary name.
	GetDefinition(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]coremodels.DeviceDefinitionTablelandModel, error)
	// QueryDefinitionsByManufacturer pages through a manufacturer's definitions, 500 at a time.
	QueryDefinitionsByManufacturer(ctx context.Context, manufacturerID int, pageIndex int) ([]coremodels.DeviceDefinitionTablelandModel, error)
	Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error)
	Update(ctx context.Context, manufacturerName string, input coremodels.DeviceDefinitionUpdateInput) (*string, error)
	Delete(ctx context.Context, manufacturerName, id string) (*string, error)
}

const (
	catalogPageSize     = 500
	manifestCacheKey    = "definitions_manifest"
	manifestCacheTTL    = time.Minute
	manufacturersCached = "manufacturers_by_token_id"
)

type catalogManufacturer struct {
	TokenID int    `json:"tokenId"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
}

type catalogDoc struct {
	coremodels.DeviceDefinitionTablelandModel
	Manufacturer catalogManufacturer `json:"manufacturer"`
}

type catalogManifest struct {
	UpdatedAt   string       `json:"updatedAt"`
	Count       int          `json:"count"`
	Definitions []catalogDoc `json:"definitions"`
}

type deviceDefinitionCatalogService struct {
	settings    *config.Settings
	logger      *zerolog.Logger
	identityAPI IdentityAPI
	httpClient  *http.Client
	memCache    *cache.Cache
}

func NewDeviceDefinitionCatalogService(settings *config.Settings, logger *zerolog.Logger) DeviceDefinitionCatalogService {
	return &deviceDefinitionCatalogService{
		settings:    settings,
		logger:      logger,
		identityAPI: NewIdentityAPIService(logger, settings),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		memCache:    cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (e *deviceDefinitionCatalogService) catalogURL(pathSuffix string) string {
	return strings.TrimSuffix(e.settings.DefinitionsCatalogURL, "/") + pathSuffix
}

func (e *deviceDefinitionCatalogService) fetchDoc(ctx context.Context, id string) (*catalogDoc, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.catalogURL("/definitions/"+id+".json"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch definition %s from catalog", id)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned %d for definition %s", resp.StatusCode, id)
	}
	var doc catalogDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, errors.Wrapf(err, "failed to decode definition %s", id)
	}
	return &doc, nil
}

func (e *deviceDefinitionCatalogService) manifest(ctx context.Context) (*catalogManifest, error) {
	if v, ok := e.memCache.Get(manifestCacheKey); ok {
		return v.(*catalogManifest), nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.catalogURL("/manifest.json"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch definitions manifest")
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned %d for manifest", resp.StatusCode)
	}
	var m catalogManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, errors.Wrap(err, "failed to decode definitions manifest")
	}
	e.memCache.Set(manifestCacheKey, &m, manifestCacheTTL)
	return &m, nil
}

func (e *deviceDefinitionCatalogService) GetManufacturer(manufacturerSlug string) (*coremodels.Manufacturer, error) {
	return e.identityAPI.GetManufacturer(manufacturerSlug)
}

func (e *deviceDefinitionCatalogService) GetManufacturerNameByID(_ context.Context, manufacturerID *big.Int) (string, error) {
	byID := map[int]string{}
	if v, ok := e.memCache.Get(manufacturersCached); ok {
		byID = v.(map[int]string)
	} else {
		all, err := e.identityAPI.GetManufacturers()
		if err != nil {
			return "", errors.Wrap(err, "failed to get manufacturers from identity")
		}
		for _, m := range all {
			byID[m.TokenID] = m.Name
		}
		e.memCache.Set(manufacturersCached, byID, 10*time.Minute)
	}
	name, ok := byID[int(manufacturerID.Int64())]
	if !ok {
		return "", fmt.Errorf("no manufacturer found for token id %d", manufacturerID.Int64())
	}
	return name, nil
}

func (e *deviceDefinitionCatalogService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error) {
	doc, err := e.fetchDoc(ctx, ID)
	if err != nil || doc == nil {
		return nil, err
	}
	if manufacturerID != nil && int64(doc.Manufacturer.TokenID) != manufacturerID.Int64() {
		return nil, nil
	}
	return &doc.DeviceDefinitionTablelandModel, nil
}

func (e *deviceDefinitionCatalogService) GetDefinition(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error) {
	return e.GetDeviceDefinitionByID(ctx, manufacturerID, ID)
}

func (e *deviceDefinitionCatalogService) GetDefinitionByID(ctx context.Context, ID string) (*coremodels.DeviceDefinitionTablelandModel, *big.Int, error) {
	doc, err := e.fetchDoc(ctx, ID)
	if err != nil || doc == nil {
		return nil, nil, err
	}
	return &doc.DeviceDefinitionTablelandModel, big.NewInt(int64(doc.Manufacturer.TokenID)), nil
}

func (e *deviceDefinitionCatalogService) GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]coremodels.DeviceDefinitionTablelandModel, error) {
	m, err := e.manifest(ctx)
	if err != nil {
		return nil, err
	}
	var manufTokenID *int64
	if !manufacturerID.IsZero() {
		v := manufacturerID.Big.Int(nil).Int64()
		manufTokenID = &v
	}
	matches := make([]coremodels.DeviceDefinitionTablelandModel, 0)
	for _, d := range m.Definitions {
		if manufTokenID != nil && int64(d.Manufacturer.TokenID) != *manufTokenID {
			continue
		}
		if ID != "" && d.ID != ID {
			continue
		}
		if model != "" && !strings.EqualFold(d.Model, model) {
			continue
		}
		if year > 0 && d.Year != year {
			continue
		}
		matches = append(matches, d.DeviceDefinitionTablelandModel)
	}
	return paginate(matches, int(pageIndex), int(pageSize)), nil
}

func (e *deviceDefinitionCatalogService) QueryDefinitionsByManufacturer(ctx context.Context, manufacturerID int, pageIndex int) ([]coremodels.DeviceDefinitionTablelandModel, error) {
	m, err := e.manifest(ctx)
	if err != nil {
		return nil, err
	}
	matches := make([]coremodels.DeviceDefinitionTablelandModel, 0)
	for _, d := range m.Definitions {
		if d.Manufacturer.TokenID == manufacturerID {
			matches = append(matches, d.DeviceDefinitionTablelandModel)
		}
	}
	return paginate(matches, pageIndex, catalogPageSize), nil
}

func paginate(items []coremodels.DeviceDefinitionTablelandModel, pageIndex, pageSize int) []coremodels.DeviceDefinitionTablelandModel {
	if pageSize <= 0 {
		pageSize = catalogPageSize
	}
	start := pageIndex * pageSize
	if start >= len(items) {
		return []coremodels.DeviceDefinitionTablelandModel{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

// workerRequest sends an authenticated request to the definitions-worker.
// When no worker URL is configured (local dev), writes are no-ops.
func (e *deviceDefinitionCatalogService) workerRequest(ctx context.Context, method, pathSuffix string, body any) (bool, error) {
	if e.settings.DefinitionsWorkerURL == "" {
		e.logger.Info().Msgf("DefinitionsWorkerURL not set, skipping %s %s", method, pathSuffix)
		return false, nil
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return false, err
		}
		reader = bytes.NewReader(b)
	}
	url := strings.TrimSuffix(e.settings.DefinitionsWorkerURL, "/") + pathSuffix
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.settings.DefinitionsWorkerToken)
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return false, errors.Wrapf(err, "definitions-worker %s %s failed", method, pathSuffix)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return false, fmt.Errorf("definitions-worker %s %s returned %d: %s", method, pathSuffix, resp.StatusCode, string(msg))
	}
	return true, nil
}

type workerPutBody struct {
	ID         string                               `json:"id"`
	Model      string                               `json:"model"`
	Year       int                                  `json:"year"`
	DeviceType string                               `json:"devicetype,omitempty"`
	ImageURI   string                               `json:"imageuri,omitempty"`
	Metadata   *coremodels.DeviceDefinitionMetadata `json:"metadata,omitempty"`
	KSUID      string                               `json:"ksuid,omitempty"`
}

func (e *deviceDefinitionCatalogService) Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error) {
	e.logger.Info().Msgf("catalog create for device definition %s (manufacturer %s)", dd.ID, manufacturerName)
	sent, err := e.workerRequest(ctx, http.MethodPut, "/definitions/"+dd.ID, workerPutBody{
		ID:         dd.ID,
		Model:      dd.Model,
		Year:       dd.Year,
		DeviceType: dd.DeviceType,
		ImageURI:   dd.ImageURI,
		Metadata:   dd.Metadata,
		KSUID:      dd.KSUID,
	})
	if err != nil || !sent {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &dd.ID, nil
}

func (e *deviceDefinitionCatalogService) Update(ctx context.Context, manufacturerName string, input coremodels.DeviceDefinitionUpdateInput) (*string, error) {
	existing, _, err := e.GetDefinitionByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("device definition %s not found in catalog to update", input.ID)
	}
	body := workerPutBody{
		ID:       input.ID,
		Model:    existing.Model,
		Year:     existing.Year,
		KSUID:    existing.KSUID,
		Metadata: existing.Metadata,
	}
	body.DeviceType = existing.DeviceType
	if input.DeviceType != "" {
		body.DeviceType = input.DeviceType
	}
	body.ImageURI = existing.ImageURI
	if input.ImageURI != "" {
		body.ImageURI = input.ImageURI
	}
	if input.Metadata != nil {
		body.Metadata = input.Metadata
	}
	sent, err := e.workerRequest(ctx, http.MethodPut, "/definitions/"+input.ID, body)
	if err != nil || !sent {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &input.ID, nil
}

func (e *deviceDefinitionCatalogService) Delete(ctx context.Context, manufacturerName, id string) (*string, error) {
	e.logger.Info().Msgf("catalog delete for device definition %s (manufacturer %s)", id, manufacturerName)
	sent, err := e.workerRequest(ctx, http.MethodDelete, "/definitions/"+id, nil)
	if err != nil || !sent {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &id, nil
}
