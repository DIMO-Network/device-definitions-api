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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/vocabulary"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
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
	// Create() is that caller: it uses this to decide whether a template
	// already exists before writing one, and must distinguish
	// ErrTemplateNotFound from every other failure. Reading through the CDN
	// there would let a stale 404 authorise a duplicate.
	GetTemplateByIDFresh(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error)
	// GetDefinition is GetDeviceDefinitionByID under its historical secondary name.
	GetDefinition(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error)
	Delete(ctx context.Context, manufacturerName, id string) (*string, error)
}

const (
	// CatalogPageSize is the page size QueryDefinitionsByManufacturer returns.
	// Exported because callers page until a short page and must agree with it;
	// a private copy elsewhere silently truncates their iteration if it drifts.
	CatalogPageSize     = 500
	manifestCacheKey    = "definitions_manifest"
	manifestCacheTTL    = time.Minute
	manufacturersCached = "manufacturers_by_token_id"
)

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

// GetDeviceDefinitionByID returns the flat pre-migration shape for the callers
// that still take it, built from the template's SHARED attributes only.
//
// Attributes that vary by trim are deliberately absent. This shape has exactly
// one slot per attribute, and filling it from an arbitrary trim is precisely
// what produced the record that claimed powertrain ICE while carrying a
// hybrid's tank size. Absent is honest; a guess is not.
//
// (nil, nil) still means "not found", as it did before, and also covers a
// template owned by a different manufacturer. Unlike GetTemplateByID this does
// not turn a missing template into an error: its callers branch on nil and
// report it, and none of them create anything on the strength of it.
func (e *deviceDefinitionCatalogService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error) {
	tmpl, tokenID, err := e.GetTemplateByID(ctx, ID)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if manufacturerID != nil && tokenID.Int64() != manufacturerID.Int64() {
		return nil, nil
	}
	return templateToDefinitionModel(tmpl), nil
}

// templateToDefinitionModel flattens a template into the legacy shape. KSUID is
// not populated: it is gone from the contract, and inventing one here would put
// a value into a field consumers read as an identifier.
func templateToDefinitionModel(tmpl *coremodels.Template) *coremodels.DeviceDefinitionTablelandModel {
	names := make([]string, 0, len(tmpl.Attributes))
	for name := range tmpl.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	attrs := make([]coremodels.DeviceTypeAttribute, 0, len(names))
	for _, name := range names {
		attrs = append(attrs, coremodels.DeviceTypeAttribute{Name: name, Value: attributeString(tmpl.Attributes[name])})
	}
	return &coremodels.DeviceDefinitionTablelandModel{
		ID:         tmpl.ID,
		Model:      tmpl.Model,
		Year:       tmpl.Year,
		DeviceType: tmpl.DeviceType,
		ImageURI:   tmpl.ImageURI,
		Metadata:   &coremodels.DeviceDefinitionMetadata{DeviceAttributes: attrs},
	}
}

// attributeString renders a typed attribute for the legacy string-valued shape.
// strconv rather than fmt so 15.8 renders "15.8" and not "1.58e+01".
func attributeString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprint(v)
	}
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

// templatePutBody is Template minus the fields definitions-worker owns.
// Marshalling coremodels.Template directly would send "version": 0, and the
// worker rejects a client-supplied version, createdAt, updatedAt or author by
// name rather than stripping it -- a writer that thinks it set them and was
// quietly overruled has learned nothing.
type templatePutBody struct {
	ID                 string                          `json:"id"`
	DeviceType         string                          `json:"deviceType"`
	Manufacturer       coremodels.TemplateManufacturer `json:"manufacturer"`
	Model              string                          `json:"model"`
	Year               int                             `json:"year"`
	ImageURI           string                          `json:"imageURI,omitempty"`
	HardwareTemplateID string                          `json:"hardwareTemplateId,omitempty"`
	Attributes         map[string]any                  `json:"attributes"`
	Trims              []templatePutTrim               `json:"trims"`
}

// templatePutTrim omits Selectors when empty. coremodels.Trim marshals them
// unconditionally, and a single-trim template has nothing to select on.
type templatePutTrim struct {
	Name               string                    `json:"name"`
	Selectors          *coremodels.TrimSelectors `json:"selectors,omitempty"`
	HardwareTemplateID string                    `json:"hardwareTemplateId,omitempty"`
	Attributes         map[string]any            `json:"attributes"`
}

// defaultTrimName matches the extraction pipeline's fallback
// (definitions-worker/scripts/extract/trims.mjs): a model-year sold in one
// configuration still has a trim, because the schema has no other shape for it.
const defaultTrimName = "Base"

const (
	vocabularyCacheKey = "device_type_vocabulary"
	// The worker serves /schema with max-age=300, so holding it longer than
	// that would show a contributor a vocabulary the validator has stopped
	// using.
	vocabularyCacheTTL = 5 * time.Minute
)

// vocabulary fetches the DeviceType the template's attributes are validated
// against. It is read from the worker rather than vendored: a stale local copy
// would fold values the validator rejects, and the mismatch would only appear
// as a failed write.
func (e *deviceDefinitionCatalogService) vocabulary(ctx context.Context, deviceType string) (*vocabulary.DeviceType, error) {
	key := vocabularyCacheKey + ":" + deviceType
	if v, ok := e.memCache.Get(key); ok {
		return v.(*vocabulary.DeviceType), nil
	}
	base := e.settings.DefinitionsWorkerURL
	if base == "" {
		base = e.settings.DefinitionsCatalogURL
	}
	reqURL := strings.TrimSuffix(base, "/") + "/schema/device-type-" + url.PathEscape(deviceType) + ".json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch the %s vocabulary", deviceType)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned %d for the %s vocabulary", resp.StatusCode, deviceType)
	}
	var vocab vocabulary.DeviceType
	if err := json.NewDecoder(resp.Body).Decode(&vocab); err != nil {
		return nil, errors.Wrapf(err, "failed to decode the %s vocabulary", deviceType)
	}
	e.memCache.Set(key, &vocab, vocabularyCacheTTL)
	return &vocab, nil
}

// templateFromDefinition builds the template for a definition that does not
// exist yet, folding dd.Metadata onto the DeviceType vocabulary.
//
// A vocabulary that cannot be fetched is an error, not a reason to write fewer
// attributes: silently storing less because a fetch failed is indistinguishable
// afterwards from the source not having carried the value.
//
// Attributes go on the template rather than the trim. There is exactly one
// trim, so every attribute is shared by all of them, and the contract requires
// an attribute to be in one place or the other and never both.
func (e *deviceDefinitionCatalogService) templateFromDefinition(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (templatePutBody, []vocabulary.DroppedAttribute, error) {
	slug := dd.ID
	if i := strings.Index(dd.ID, "_"); i > 0 {
		slug = dd.ID[:i]
	}
	manufacturer := coremodels.TemplateManufacturer{Slug: slug, Name: manufacturerName}
	// Best effort: tokenId is optional in the contract precisely so that a
	// template can be created for a brand with no on-chain storage yet.
	if m, err := e.GetManufacturer(slug); err == nil && m != nil {
		manufacturer.TokenID = m.TokenID
		if manufacturer.Name == "" {
			manufacturer.Name = m.Name
		}
	}
	deviceType := dd.DeviceType
	if deviceType == "" {
		deviceType = common.DefaultDeviceType
	}

	vocab, err := e.vocabulary(ctx, deviceType)
	if err != nil {
		return templatePutBody{}, nil, err
	}
	raw := map[string]string{}
	if dd.Metadata != nil {
		for _, a := range dd.Metadata.DeviceAttributes {
			raw[a.Name] = a.Value
		}
	}
	attributes, dropped := vocabulary.MapAttributes(raw, vocab)

	return templatePutBody{
		ID:           dd.ID,
		DeviceType:   deviceType,
		Manufacturer: manufacturer,
		Model:        dd.Model,
		Year:         dd.Year,
		ImageURI:     dd.ImageURI,
		Attributes:   attributes,
		Trims:        []templatePutTrim{{Name: defaultTrimName, Attributes: map[string]any{}}},
	}, dropped, nil
}

func (e *deviceDefinitionCatalogService) Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error) {
	e.logger.Info().Msgf("catalog create for device definition %s (manufacturer %s)", dd.ID, manufacturerName)
	// The worker PUT is an upsert; preserve the old create semantics so a
	// create can never silently overwrite a curated template. Fresh read so a
	// stale CDN 404 can't slip through. Only ErrTemplateNotFound means "does
	// not exist" -- every other error must abort, or a catalog outage reads as
	// a green light to create a duplicate.
	_, _, err := e.GetTemplateByIDFresh(ctx, dd.ID)
	if err == nil {
		return nil, fmt.Errorf("cannot create device definition, already exists: %s", dd.ID)
	}
	if !errors.Is(err, ErrTemplateNotFound) {
		return nil, err
	}
	body, dropped, err := e.templateFromDefinition(ctx, manufacturerName, dd)
	if err != nil {
		return nil, err
	}
	// Report what was read and not carried, not only what failed to parse. A
	// value dropped without a trace is the defect class this migration keeps
	// producing.
	for _, d := range dropped {
		e.logger.Info().Msgf("catalog create %s: dropped %s", dd.ID, d)
	}
	if _, err := e.workerRequest(ctx, http.MethodPut, "/t/"+url.PathEscape(dd.ID), body); err != nil {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &dd.ID, nil
}

func (e *deviceDefinitionCatalogService) Delete(ctx context.Context, manufacturerName, id string) (*string, error) {
	e.logger.Info().Msgf("catalog delete for device definition %s (manufacturer %s)", id, manufacturerName)
	if _, err := e.workerRequest(ctx, http.MethodDelete, "/t/"+url.PathEscape(id), nil); err != nil {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &id, nil
}
