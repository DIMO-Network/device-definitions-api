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
	"net/url"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Metric method labels for the shared request counters; dashboards filter on
// these the way they previously filtered on the Tableland* labels.
const (
	metricCatalogRead  = "CatalogRead"
	metricManifestRead = "CatalogManifestRead"
	metricCatalogWrite = "CatalogWrite"
)

// ErrTemplateNotFound is returned by GetTemplateByID/GetTemplateByIDFresh
// only for a genuine 404 from t/<id>.json -- the template import has not run
// for this id. Callers must check it with errors.Is, never by inspecting the
// error's message: any other status, a transport error, or a decode failure
// returns a distinct, non-sentinel error, so a catalog outage is never
// mistaken for "this vehicle does not exist." Conflating the two would let a
// caller respond to a 500 by creating a duplicate definition mid-outage.
var ErrTemplateNotFound = errors.New("template not found in catalog")

func countOutcome(method string, err error) {
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": method}).Inc()
		return
	}
	metrics.Success.With(prometheus.Labels{"method": method}).Inc()
}

//go:generate mockgen -source device_definition_catalog_service.go -destination mocks/device_definition_catalog_service_mock.go -package mocks

// DeviceDefinitionCatalogService reads device definitions from the R2-backed
// catalog (CDN) and writes them through the definitions-worker. It replaces
// the Tableland on-chain service.
type DeviceDefinitionCatalogService interface {
	GetManufacturer(manufacturerSlug string) (*coremodels.Manufacturer, error)
	GetManufacturerNameByID(ctx context.Context, manufacturerID *big.Int) (string, error)
	// GetDeviceDefinitionByID gets a definition by slug ID, requiring it to belong to the given manufacturer.
	GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	// GetTemplateByID reads the vehicle template for a slug ID from t/<id>.json
	// and returns the manufacturer token id too. It does not fall back to the
	// pre-migration definitions/<id>.json: a 404 here means the template
	// import has not run for this id, and that must fail loudly rather than
	// silently serving the flat pre-migration record.
	GetTemplateByID(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error)
	// GetTemplateByIDFresh is GetTemplateByID reading through the worker
	// instead of the CDN. A caller that reads, mutates and writes back must use
	// it: documents are served with max-age=86400, so a cached read silently
	// discards any edit made in the last day when the result is PUT back.
	//
	// It has NO production caller today. Its only one was the bulk powertrain
	// tool, deleted with the legacy Update() write path earlier on this branch.
	// It is kept because Create() is still on the pre-migration
	// /definitions/<id> route the new worker does not serve, and migrating that
	// write to templates is the read-modify-write this method exists for. If
	// that migration lands without using it, delete it -- do not leave it here
	// on the strength of this comment alone.
	GetTemplateByIDFresh(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error)
	// GetDefinition is GetDeviceDefinitionByID under its historical secondary name.
	GetDefinition(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]coremodels.DeviceDefinitionTablelandModel, error)
	// QueryDefinitionsByManufacturer pages through a manufacturer's definitions, 500 at a time.
	QueryDefinitionsByManufacturer(ctx context.Context, manufacturerID int, pageIndex int) ([]coremodels.DeviceDefinitionTablelandModel, error)
	// CatalogIDs returns every definition id in the catalog, indexed or not.
	// Deletion decisions must be based on this rather than on walking
	// manufacturers: the manifest is the catalog, whereas a per-manufacturer
	// walk is only as complete as identity-api's manufacturer list.
	CatalogIDs(ctx context.Context) ([]string, error)
	// PinCatalogSnapshot fetches the manifest once and holds it for the caller's
	// whole run. Anything paging over the catalog to decide what to delete must
	// call it first: the per-request cache expires after a minute while a full
	// sync takes several, so without pinning, pages come from different
	// manifests and a definition that shifts position between them is returned
	// by neither -- and is then deleted as an orphan.
	//
	// It is not a cache bypass. The worker generates every response for
	// definitions.dimo.org (no cf-cache-status, no age, date advances per
	// request), so there is no edge cache in front of the manifest.
	PinCatalogSnapshot(ctx context.Context) error
	Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error)
	Delete(ctx context.Context, manufacturerName, id string) (*string, error)
}

const (
	// CatalogPageSize is the page size QueryDefinitionsByManufacturer returns.
	// Exported because callers page until a short page and must agree with it;
	// a private copy elsewhere silently truncates their iteration if it drifts.
	CatalogPageSize  = 500
	manifestCacheKey = "definitions_manifest"
	manifestCacheTTL = time.Minute
	// Long enough to cover a full sync run, which pages over the whole catalog.
	pinnedManifestTTL   = time.Hour
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

// UnmarshalJSON exists because the embedded model's custom UnmarshalJSON
// would otherwise be promoted to catalogDoc and silently drop Manufacturer.
func (d *catalogDoc) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &d.DeviceDefinitionTablelandModel); err != nil {
		return err
	}
	var aux struct {
		Manufacturer catalogManufacturer `json:"manufacturer"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	d.Manufacturer = aux.Manufacturer
	return nil
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
	doc, err := e.fetchDocFrom(ctx, e.catalogURL("/definitions/"+url.PathEscape(id)+".json"), id)
	countOutcome(metricCatalogRead, err)
	return doc, err
}

// fetchDocFresh bypasses the CDN by reading through the worker when
// configured, so read-modify-write never merges a stale cached base.
func (e *deviceDefinitionCatalogService) fetchDocFresh(ctx context.Context, id string) (*catalogDoc, error) {
	base := e.settings.DefinitionsWorkerURL
	if base == "" {
		return e.fetchDoc(ctx, id)
	}
	doc, err := e.fetchDocFrom(ctx, strings.TrimSuffix(base, "/")+"/definitions/"+url.PathEscape(id), id)
	countOutcome(metricCatalogRead, err)
	return doc, err
}

func (e *deviceDefinitionCatalogService) fetchDocFrom(ctx context.Context, reqURL, id string) (*catalogDoc, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

// fetchTemplateDoc reads the vehicle template for id from t/<id>.json. There
// is deliberately no fallback to fetchDoc's definitions/<id>.json: falling
// back would serve the pre-migration flat record and hide an incomplete
// template import behind decodes that appear to work.
func (e *deviceDefinitionCatalogService) fetchTemplateDoc(ctx context.Context, id string) (*coremodels.Template, error) {
	tmpl, err := e.fetchTemplateDocFrom(ctx, e.catalogURL("/t/"+url.PathEscape(id)+".json"), id)
	countOutcome(metricCatalogRead, err)
	return tmpl, err
}

// fetchTemplateDocFresh bypasses the CDN by reading through the worker when
// configured, so read-modify-write never merges a stale cached base.
func (e *deviceDefinitionCatalogService) fetchTemplateDocFresh(ctx context.Context, id string) (*coremodels.Template, error) {
	base := e.settings.DefinitionsWorkerURL
	if base == "" {
		return e.fetchTemplateDoc(ctx, id)
	}
	tmpl, err := e.fetchTemplateDocFrom(ctx, strings.TrimSuffix(base, "/")+"/t/"+url.PathEscape(id)+".json", id)
	countOutcome(metricCatalogRead, err)
	return tmpl, err
}

func (e *deviceDefinitionCatalogService) fetchTemplateDocFrom(ctx context.Context, reqURL, id string) (*coremodels.Template, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch template %s from catalog", id)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode == http.StatusNotFound {
		// Unlike fetchDocFrom, a 404 is an error here, not a nil result: it
		// means the template import for this id has not run, and that must
		// fail loudly rather than be treated as a normal "not found". It is
		// ErrTemplateNotFound specifically -- and only this -- so a caller
		// can tell "does not exist" apart from every other failure below.
		return nil, errors.Wrapf(ErrTemplateNotFound, "template %s", id)
	}
	if resp.StatusCode != http.StatusOK {
		// Any other status is a catalog problem, not a missing template:
		// callers must not reclassify this as not-found.
		return nil, fmt.Errorf("catalog returned %d for template %s", resp.StatusCode, id)
	}
	var tmpl coremodels.Template
	if err := json.NewDecoder(resp.Body).Decode(&tmpl); err != nil {
		return nil, errors.Wrapf(err, "failed to decode template %s", id)
	}
	return &tmpl, nil
}

func (e *deviceDefinitionCatalogService) manifest(ctx context.Context) (*catalogManifest, error) {
	if v, ok := e.memCache.Get(manifestCacheKey); ok {
		return v.(*catalogManifest), nil
	}
	// Every exit counts itself. Previously only a decode failure incremented the
	// error metric, so a catalog outage showed up as these series dropping to
	// zero rather than as an error spike -- invisible to any alert written as
	// "error rate > X".
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.catalogURL("/manifest.json"), nil)
	if err != nil {
		countOutcome(metricManifestRead, err)
		return nil, err
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		countOutcome(metricManifestRead, err)
		return nil, errors.Wrap(err, "failed to fetch definitions manifest")
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("catalog returned %d for manifest", resp.StatusCode)
		countOutcome(metricManifestRead, err)
		return nil, err
	}
	var m catalogManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		countOutcome(metricManifestRead, err)
		return nil, errors.Wrap(err, "failed to decode definitions manifest")
	}
	countOutcome(metricManifestRead, nil)
	e.memCache.Set(manifestCacheKey, &m, manifestCacheTTL)
	return &m, nil
}

func (e *deviceDefinitionCatalogService) CatalogIDs(ctx context.Context) ([]string, error) {
	m, err := e.manifest(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(m.Definitions))
	for _, d := range m.Definitions {
		ids = append(ids, d.ID)
	}
	return ids, nil
}

func (e *deviceDefinitionCatalogService) PinCatalogSnapshot(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.catalogURL("/manifest.json"), nil)
	if err != nil {
		countOutcome(metricManifestRead, err)
		return err
	}
	// Cheap insurance if an edge cache is ever put in front of the catalog.
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		countOutcome(metricManifestRead, err)
		return errors.Wrap(err, "failed to fetch a fresh definitions manifest")
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("catalog returned %d for a fresh manifest", resp.StatusCode)
		countOutcome(metricManifestRead, err)
		return err
	}
	var m catalogManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		countOutcome(metricManifestRead, err)
		return errors.Wrap(err, "failed to decode the fresh definitions manifest")
	}
	if m.Count != len(m.Definitions) {
		err := fmt.Errorf("manifest count %d does not match the %d definitions it carries", m.Count, len(m.Definitions))
		countOutcome(metricManifestRead, err)
		return err
	}
	countOutcome(metricManifestRead, nil)
	// Long TTL so every page of the caller's run reads the same snapshot.
	e.memCache.Set(manifestCacheKey, &m, pinnedManifestTTL)
	return nil
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
		// identity answers 200 with an `errors` array and null data when it
		// fails, which the client surfaces as (nil, nil). Caching that empty
		// map for ten minutes turns a brief blip into ten minutes of every
		// lookup failing, and callers that treat the failure as "skip this row"
		// drop work silently.
		if len(all) == 0 {
			return "", fmt.Errorf("identity returned no manufacturers")
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

func (e *deviceDefinitionCatalogService) GetTemplateByID(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error) {
	tmpl, err := e.fetchTemplateDoc(ctx, ID)
	if err != nil {
		return nil, nil, err
	}
	return tmpl, big.NewInt(int64(tmpl.Manufacturer.TokenID)), nil
}

func (e *deviceDefinitionCatalogService) GetTemplateByIDFresh(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error) {
	tmpl, err := e.fetchTemplateDocFresh(ctx, ID)
	if err != nil {
		return nil, nil, err
	}
	return tmpl, big.NewInt(int64(tmpl.Manufacturer.TokenID)), nil
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
	return paginate(matches, pageIndex, CatalogPageSize), nil
}

func paginate(items []coremodels.DeviceDefinitionTablelandModel, pageIndex, pageSize int) []coremodels.DeviceDefinitionTablelandModel {
	if pageSize <= 0 {
		pageSize = CatalogPageSize
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
// An unconfigured worker URL is an error, not a mode: silently skipping the
// write and reporting success meant a decode could answer 200 with a definition
// id that was never written to R2, and a delete could log success for a
// definition that still exists.
func (e *deviceDefinitionCatalogService) workerRequest(ctx context.Context, method, pathSuffix string, body any) (bool, error) {
	if e.settings.DefinitionsWorkerURL == "" {
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		return false, fmt.Errorf("definitions-worker is not configured (DEFINITIONS_WORKER_URL is empty); refusing to report %s %s as written", method, pathSuffix)
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return false, err
		}
		reader = bytes.NewReader(b)
	}
	reqURL := strings.TrimSuffix(e.settings.DefinitionsWorkerURL, "/") + pathSuffix
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.settings.DefinitionsWorkerToken)
	resp, err := e.httpClient.Do(req)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		return false, errors.Wrapf(err, "definitions-worker %s %s failed", method, pathSuffix)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode >= 300 {
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return false, fmt.Errorf("definitions-worker %s %s returned %d: %s", method, pathSuffix, resp.StatusCode, string(msg))
	}
	metrics.Success.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
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
	// The worker PUT is an upsert; preserve the old create semantics so a
	// create can never silently overwrite a curated definition. Fresh read so
	// a stale CDN 404 can't slip through.
	existing, err := e.fetchDocFresh(ctx, dd.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("cannot create device definition, already exists: %s", dd.ID)
	}
	_, err = e.workerRequest(ctx, http.MethodPut, "/definitions/"+url.PathEscape(dd.ID), workerPutBody{
		ID:         dd.ID,
		Model:      dd.Model,
		Year:       dd.Year,
		DeviceType: dd.DeviceType,
		ImageURI:   dd.ImageURI,
		Metadata:   dd.Metadata,
		KSUID:      dd.KSUID,
	})
	if err != nil {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &dd.ID, nil
}

func (e *deviceDefinitionCatalogService) Delete(ctx context.Context, manufacturerName, id string) (*string, error) {
	e.logger.Info().Msgf("catalog delete for device definition %s (manufacturer %s)", id, manufacturerName)
	if _, err := e.workerRequest(ctx, http.MethodDelete, "/definitions/"+url.PathEscape(id), nil); err != nil {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &id, nil
}
