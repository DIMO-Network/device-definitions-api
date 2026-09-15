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
	"sync/atomic"
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

// ErrManufacturerUnresolved is returned by Create when identity-api cannot
// resolve the definition's manufacturer to a minted token id, whether because
// the make is not minted or because identity failed to answer. Nothing is
// written. A template created without a tokenId cannot be used by
// UpsertDecoding or addvin, and the on-chain create this replaced refused the
// same case.
var ErrManufacturerUnresolved = errors.New("manufacturer not resolved in identity")

// ErrTemplateExists is returned by Create when a template is already stored at
// the id: the worker answered 412 to its If-None-Match: * write, and nothing
// was written. A caller that only needs the template to exist -- the decode
// path -- treats it as success by another writer and reads the stored template
// back; it must not retry the write unconditionally.
var ErrTemplateExists = errors.New("template already exists in catalog")

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
	// and returns the manufacturer token id too. The token id is nil when the
	// template carries none: tokenId is optional in the contract, and a caller
	// must treat nil as unknown, never as manufacturer 0. It does not fall
	// back to the pre-migration definitions/<id>.json: a 404 here means the template
	// import has not run for this id, and that must fail loudly rather than
	// silently serving the flat pre-migration record.
	GetTemplateByID(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error)
	// GetTemplateByIDFresh is GetTemplateByID with a request the edge cache
	// cannot answer. A caller that reads, mutates and writes back must use it:
	// documents are served with max-age=86400, so a cached read silently
	// discards any edit made in the last day when the result is PUT back.
	//
	// The worker and the catalog are one hostname, so this is not a different
	// host -- it is a different cache key, a unique query parameter the worker
	// ignores. See fetchTemplateDocFresh.
	//
	// The decode path uses it after Create returns ErrTemplateExists, to read
	// back the template another writer just stored: the edge may still be
	// serving the 404 the decode saw moments earlier.
	GetTemplateByIDFresh(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error)
	// GetDefinition is GetDeviceDefinitionByID under its historical secondary name.
	GetDefinition(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	// Create writes a template that does not exist yet and returns the
	// document the worker stored -- the PUT's own response, so no read-back
	// is needed. A caller must narrow that document to a trim exactly as it
	// would a template it had found, or a definition's first decode answers
	// differently from its second.
	Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*coremodels.Template, error)
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

// freshParam is the query parameter that makes a fresh read miss the edge
// cache. definitions-worker routes on url.pathname alone (src/index.ts's
// templateMatch), so it never reads this and the request reaches exactly the
// same handler the cached read reaches; Cloudflare's cache key is the whole
// URL, so a value it has not seen cannot be served from cache.
const freshParam = "fresh"

// freshSeq disambiguates two fresh reads minted inside one clock tick. The
// value only has to be unique, never unguessable.
var freshSeq atomic.Uint64

func freshValue() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.FormatUint(freshSeq.Add(1), 36)
}

// fetchTemplateDocFresh reads a template past the edge cache, so
// read-modify-write never merges a stale cached base and a read-back after a
// create-only conflict is never served the 404 the create just saw.
//
// It does NOT read "through the worker instead of the CDN": there is no second
// host to read through. DEFINITIONS_CATALOG_URL and DEFINITIONS_WORKER_URL are
// the same hostname in both dev and prod (charts/device-definitions-api),
// because the worker IS what serves definitions.dimo.org -- so the fresh read
// built a byte-identical request to the cached one and got the cached answer.
// Keeping one hostname and missing the cache by cache key is the decision:
// a unique query parameter the worker ignores and the edge cannot match.
//
// A request header cannot do this job: Cloudflare does not honour a client's
// Cache-Control on a cached response, and the document is served
// `public, max-age=86400, stale-while-revalidate=604800`.
func (e *deviceDefinitionCatalogService) fetchTemplateDocFresh(ctx context.Context, id string) (*coremodels.Template, error) {
	base := e.settings.DefinitionsWorkerURL
	if base == "" {
		base = e.settings.DefinitionsCatalogURL
	}
	reqURL := strings.TrimSuffix(base, "/") + "/t/" + url.PathEscape(id) + ".json?" + freshParam + "=" + freshValue()
	tmpl, err := e.fetchTemplateDocFrom(ctx, reqURL, id)
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
	// A template's tokenId is optional, so a caller can hold none. No
	// manufacturer has a token id below 1: asking identity would only fail
	// after a round trip, and a nil id would panic below.
	if manufacturerID == nil || manufacturerID.Sign() <= 0 {
		return "", fmt.Errorf("no manufacturer token id to look up: %v", manufacturerID)
	}
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
	// Ownership is compared only when both sides are known. A template with no
	// tokenId belongs to no token, so it cannot belong to another manufacturer's.
	if manufacturerID != nil && tokenID != nil && tokenID.Int64() != manufacturerID.Int64() {
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
	return tmpl, manufacturerTokenID(tmpl), nil
}

func (e *deviceDefinitionCatalogService) GetTemplateByIDFresh(ctx context.Context, ID string) (*coremodels.Template, *big.Int, error) {
	tmpl, err := e.fetchTemplateDocFresh(ctx, ID)
	if err != nil {
		return nil, nil, err
	}
	return tmpl, manufacturerTokenID(tmpl), nil
}

// manufacturerTokenID is the template's manufacturer token id, or nil when the
// template carries none. tokenId is optional in the contract and at least 1
// when present, so the zero value can only mean absent. Reporting it as token
// id 0 sent callers to look up, and compare against, a manufacturer that does
// not exist.
func manufacturerTokenID(tmpl *coremodels.Template) *big.Int {
	if tmpl.Manufacturer.TokenID <= 0 {
		return nil
	}
	return big.NewInt(int64(tmpl.Manufacturer.TokenID))
}

// workerRequest sends an authenticated request to the definitions-worker.
// An unconfigured worker URL is an error, not a mode: silently skipping the
// write and reporting success meant a decode could answer 200 with a definition
// id that was never written to R2, and a delete could log success for a
// definition that still exists.
//
// It returns the worker's status code alongside any error, so a caller can act
// on a refusal it understands without reading the message; 0 means no response
// was received.
func (e *deviceDefinitionCatalogService) workerRequest(ctx context.Context, method, pathSuffix string, body any) (int, error) {
	status, _, err := e.workerRequestWithHeader(ctx, method, pathSuffix, body, nil)
	return status, err
}

// maxWorkerResponseBytes bounds the response body read back from a write. The
// worker refuses a request body over 64KB (MAX_DOC_BYTES) and answers a write
// with the document it stored, so this is generous by an order of magnitude
// and exists only so a misbehaving upstream cannot be read without limit.
const maxWorkerResponseBytes = 1 << 20

// workerRequestWithHeader is workerRequest with extra request headers, such as
// a write precondition, and it returns the response body.
//
// The worker answers a write with the document it stored -- version, author
// and timestamps stamped. A caller that needs that document must not have to
// fetch it again: the read-back is a second round trip against a CDN that may
// still be serving what the write just replaced.
func (e *deviceDefinitionCatalogService) workerRequestWithHeader(ctx context.Context, method, pathSuffix string, body any, header http.Header) (int, []byte, error) {
	if e.settings.DefinitionsWorkerURL == "" {
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		return 0, nil, fmt.Errorf("definitions-worker is not configured (DEFINITIONS_WORKER_URL is empty); refusing to report %s %s as written", method, pathSuffix)
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(b)
	}
	reqURL := strings.TrimSuffix(e.settings.DefinitionsWorkerURL, "/") + pathSuffix
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.settings.DefinitionsWorkerToken)
	for name, values := range header {
		for _, v := range values {
			req.Header.Add(name, v)
		}
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		return 0, nil, errors.Wrapf(err, "definitions-worker %s %s failed", method, pathSuffix)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode >= 300 {
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return resp.StatusCode, nil, fmt.Errorf("definitions-worker %s %s returned %d: %s", method, pathSuffix, resp.StatusCode, string(msg))
	}
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxWorkerResponseBytes))
	if err != nil {
		// The write landed; only the answer was lost. Report the status and
		// the read failure and let the caller decide -- treating this as a
		// failed write would have the decode path create a duplicate.
		metrics.InternalError.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
		return resp.StatusCode, nil, errors.Wrapf(err, "definitions-worker %s %s response could not be read", method, pathSuffix)
	}
	metrics.Success.With(prometheus.Labels{"method": metricCatalogWrite}).Inc()
	return resp.StatusCode, respBody, nil
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
// The manufacturer is resolved through identity first, and a make identity
// cannot resolve to a token id fails with ErrManufacturerUnresolved before
// anything else is fetched or written.
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
	// Not best effort. tokenId is optional in the contract, but a template
	// written without one is unusable by UpsertDecoding and addvin, and best
	// effort turned a single identity blip into a permanently tokenless
	// template. The on-chain create this replaced refused the same case.
	m, err := e.GetManufacturer(slug)
	if err != nil {
		return templatePutBody{}, nil, fmt.Errorf("%w: %s: %w", ErrManufacturerUnresolved, slug, err)
	}
	if m == nil || m.TokenID <= 0 {
		return templatePutBody{}, nil, fmt.Errorf("%w: %s has no manufacturer token id", ErrManufacturerUnresolved, slug)
	}
	manufacturer := coremodels.TemplateManufacturer{Slug: slug, Name: manufacturerName, TokenID: m.TokenID}
	if manufacturer.Name == "" {
		manufacturer.Name = m.Name
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

// Create writes the template for a definition that does not exist yet, and
// never overwrites one. The worker's PUT is an upsert unless told otherwise, so
// the write carries If-None-Match: * and the existence check happens at the
// worker, atomically with the write. Reading first and then writing
// unconditionally left a window in which a curator's save was replaced by a
// single Base trim while the create reported success.
//
// A template already stored at the id returns an error wrapping
// ErrTemplateExists, with nothing written. Every other refusal -- a 422 for an
// invalid template, an outage -- is a plain error.
//
// On success it returns the document the worker stored, which the worker
// answers the PUT with: version stamped, and the Base trim the decode has to
// narrow to. Discarding it left the decode that created a definition with no
// template to match against, so it answered an empty trim and version zero
// while the next decode of the same VIN answered Base and version 1.
//
// A response that cannot be read back as this template is (nil, nil): the
// write did land, and reporting it as failed would have the caller create a
// duplicate, but there is nothing to hand back and a caller that needs the
// document must read it. It is never (nil, nil) in normal operation.
func (e *deviceDefinitionCatalogService) Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*coremodels.Template, error) {
	e.logger.Info().Msgf("catalog create for device definition %s (manufacturer %s)", dd.ID, manufacturerName)
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
	createOnly := http.Header{"If-None-Match": []string{"*"}}
	status, respBody, err := e.workerRequestWithHeader(ctx, http.MethodPut, "/t/"+url.PathEscape(dd.ID), body, createOnly)
	if status == http.StatusPreconditionFailed {
		return nil, fmt.Errorf("%w: %s: %w", ErrTemplateExists, dd.ID, err)
	}
	if err != nil {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	var stored coremodels.Template
	if err := json.Unmarshal(respBody, &stored); err != nil {
		e.logger.Warn().Err(err).Msgf("catalog create %s stored the template but its response could not be read as one", dd.ID)
		return nil, nil
	}
	if stored.ID != dd.ID {
		// Not the document we asked for. Handing it back would have a caller
		// narrow trims for a different definition than the one it decoded.
		e.logger.Warn().Msgf("catalog create %s answered with template %q", dd.ID, stored.ID)
		return nil, nil
	}
	return &stored, nil
}

func (e *deviceDefinitionCatalogService) Delete(ctx context.Context, manufacturerName, id string) (*string, error) {
	e.logger.Info().Msgf("catalog delete for device definition %s (manufacturer %s)", id, manufacturerName)
	if _, err := e.workerRequest(ctx, http.MethodDelete, "/t/"+url.PathEscape(id), nil); err != nil {
		return nil, err
	}
	e.memCache.Delete(manifestCacheKey)
	return &id, nil
}
