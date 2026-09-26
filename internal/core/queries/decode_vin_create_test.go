package queries

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// The decode's create is only observable at the worker, so these run the real
// catalog gateway against a fake one. DecodeVINQueryHandler's suite needs
// Postgres; createOrAdoptTemplate does not.

// A curator's template: two trims, at version 2.
const curatedTemplateJSON = `{
  "id": "toyota_camry_2026",
  "deviceType": "vehicle",
  "manufacturer": {"slug": "toyota", "name": "Toyota", "tokenId": 131},
  "model": "Camry",
  "year": 2026,
  "attributes": {"number_of_doors": 4},
  "trims": [
    {"name": "LE", "selectors": {"styleName": ["LE"]}, "attributes": {"powertrain_type": "ICE"}},
    {"name": "Hybrid LE", "selectors": {"styleName": ["Hybrid LE"]}, "attributes": {"powertrain_type": "HEV"}}
  ],
  "version": 2
}`

var firstDecode = coremodels.DeviceDefinitionTablelandModel{
	ID:         "toyota_camry_2026",
	Model:      "Camry",
	Year:       2026,
	DeviceType: "vehicle",
}

// fakeWorker keeps the definitions-worker's write contract: GET /t/<id> serves
// the stored template or 404, a PUT with If-None-Match: * answers 412 when one
// is stored, and a PUT without it overwrites.
type fakeWorker struct {
	stored       string
	rejectPut    int
	puts         []string // the If-None-Match header each PUT carried
	putPaths     []string // the path each PUT went to
	templateGets int
	getPaths     []string // the path each template GET went to
	getQueries   []string // the query string each template GET carried
	edgeHits     int      // template GETs the edge answered without the origin
	cdnRequests  int
	version      int
}

// nextVersion is the version the worker stamps on the document it stores.
func (fw *fakeWorker) nextVersion() int {
	fw.version++
	return fw.version
}

// workerHandler is definitions-worker's contract: GET /t/<id> serves the
// stored template or 404, a PUT with If-None-Match: * answers 412 when one is
// stored, and a PUT without it overwrites. Routing is on the path alone
// (src/index.ts matches url.pathname), so a query string reaches nothing here
// -- which is what makes a cache-busting parameter safe.
func workerHandler(fw *fakeWorker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			// identity-api's GraphQL endpoint.
			_, _ = w.Write([]byte(`{"data":{"manufacturer":{"tokenId":131,"name":"Toyota"}}}`))
		case strings.HasPrefix(r.URL.Path, "/schema/"):
			_, _ = w.Write([]byte(`{"id":"vehicle","attributes":[]}`))
		case r.Method == http.MethodPut:
			fw.puts = append(fw.puts, r.Header.Get("If-None-Match"))
			fw.putPaths = append(fw.putPaths, r.URL.Path)
			if fw.rejectPut != 0 {
				w.WriteHeader(fw.rejectPut)
				_, _ = w.Write([]byte(`{"errors":["rejected"]}`))
				return
			}
			if r.Header.Get("If-None-Match") == "*" && fw.stored != "" {
				w.WriteHeader(http.StatusPreconditionFailed)
				_, _ = w.Write([]byte(`{"error":"template already exists","expected":null,"actual":2}`))
				return
			}
			// The worker answers a PUT with the document it stored, version
			// stamped. That body is what the decode matches trims against.
			body, _ := io.ReadAll(r.Body)
			stored, _ := sjson.SetBytes(body, "version", fw.nextVersion())
			fw.stored = string(stored)
			_, _ = w.Write(stored)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/t/"):
			fw.templateGets++
			fw.getPaths = append(fw.getPaths, r.URL.Path)
			fw.getQueries = append(fw.getQueries, r.URL.RawQuery)
			if fw.stored == "" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(fw.stored))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// edgeCache is Cloudflare in front of the worker. Template documents are
// served `public, max-age=86400`, and the cache key is the whole URL -- path
// AND query -- so a GET whose exact URL has been answered before is served
// from cache for up to a day, and only a URL the edge has not seen reaches the
// origin at all.
func edgeCache(fw *fakeWorker, origin http.Handler) http.HandlerFunc {
	type entry struct {
		status int
		body   string
	}
	cached := map[string]entry{}
	var mu sync.Mutex
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/t/") {
			origin.ServeHTTP(w, r)
			return
		}
		key := r.URL.RequestURI()
		mu.Lock()
		hit, ok := cached[key]
		if !ok {
			mu.Unlock()
			rec := httptest.NewRecorder()
			origin.ServeHTTP(rec, r)
			hit = entry{status: rec.Code, body: rec.Body.String()}
			mu.Lock()
			cached[key] = hit
		} else {
			fw.edgeHits++
		}
		mu.Unlock()
		w.WriteHeader(hit.status)
		_, _ = w.Write([]byte(hit.body))
	}
}

// newFakeWorker points the catalog and worker settings at two different
// servers, so a request that reached the worker is unambiguous in a test. The
// charts do NOT produce this shape -- see newOneHostFakeWorker, which does.
func newFakeWorker(t *testing.T, fw *fakeWorker) gateways.DeviceDefinitionCatalogService {
	t.Helper()
	// The CDN still serves the 404 the decode's read saw.
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fw.cdnRequests++
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(cdn.Close)
	worker := httptest.NewServer(workerHandler(fw))
	t.Cleanup(worker.Close)
	return catalogService(t, cdn.URL, worker.URL)
}

// newOneHostFakeWorker is the configuration the charts actually deploy:
// DEFINITIONS_CATALOG_URL and DEFINITIONS_WORKER_URL are the same hostname in
// both dev and prod (values.yaml, values-prod.yaml), with Cloudflare's cache
// in front of it. There is no second host to read "through the worker"
// instead, so the only thing that can separate a fresh read from a cached one
// is the cache key.
func newOneHostFakeWorker(t *testing.T, fw *fakeWorker) gateways.DeviceDefinitionCatalogService {
	t.Helper()
	host := httptest.NewServer(edgeCache(fw, workerHandler(fw)))
	t.Cleanup(host.Close)
	return catalogService(t, host.URL, host.URL)
}

func catalogService(t *testing.T, catalogURL, workerURL string) gateways.DeviceDefinitionCatalogService {
	t.Helper()
	identity, err := url.Parse(workerURL)
	require.NoError(t, err)
	logger := zerolog.Nop()
	return gateways.NewDeviceDefinitionCatalogService(&config.Settings{
		DefinitionsCatalogURL:  catalogURL,
		DefinitionsWorkerURL:   workerURL,
		DefinitionsWorkerToken: "test-token",
		IdentityAPIURL:         *identity,
	}, &logger)
}

// A curator saved toyota_camry_2026 with its trims after a first decode of the
// same model-year read 404 and before it created the template. The create used
// to overwrite them with a single Base trim and report success.
func TestCreateOrAdoptTemplateContinuesWithTheTemplateAnotherWriterStored(t *testing.T) {
	fw := &fakeWorker{stored: curatedTemplateJSON}
	catalog := newFakeWorker(t, fw)

	tmpl, err := createOrAdoptTemplate(context.Background(), catalog, "Toyota", firstDecode)
	require.NoError(t, err, "another writer creating the template is success, not a failed decode")
	require.NotNil(t, tmpl, "the decode must continue with the stored template")

	assert.Equal(t, []string{"*"}, fw.puts, "one create-only PUT, and no second write")
	assert.JSONEq(t, curatedTemplateJSON, fw.stored, "the curated template must not be overwritten")
	assert.Equal(t, 1, fw.templateGets, "the stored template is read back with a cache-busting read")
	assert.Zero(t, fw.cdnRequests, "the edge may still be serving the 404")

	assert.Equal(t, 2, tmpl.Version)
	require.Len(t, tmpl.Trims, 2, "the curated trims, not a single Base trim")
	assert.Equal(t, "Hybrid LE", tmpl.Trims[1].Name)
}

// When this decode's create is the one that lands, it must still answer with
// the template it wrote. Returning nothing skipped the whole match block, so
// the very first decode of a definition answered trim "", template_version 0
// and match_quality "" -- and the next decode of the same VIN, reading the
// stored Base trim, answered trim "Base" with quality "exact". Two calls a
// second apart disagreeing about one VIN is exactly what the cached path's
// doc comment says must not happen, and the trim-match counter under-reported
// every new definition.
//
// The worker's PUT already answers with the stored document, so there is
// nothing extra to fetch: no read-back, one write.
func TestCreateOrAdoptTemplateReturnsTheTemplateItWrote(t *testing.T) {
	fw := &fakeWorker{}
	catalog := newFakeWorker(t, fw)

	tmpl, err := createOrAdoptTemplate(context.Background(), catalog, "Toyota", firstDecode)
	require.NoError(t, err)
	require.NotNil(t, tmpl, "the decode has to narrow the template it just created")
	assert.Equal(t, []string{"*"}, fw.puts)
	assert.NotEmpty(t, fw.stored)
	assert.Zero(t, fw.templateGets, "the PUT's own response is the stored template")

	assert.Equal(t, "toyota_camry_2026", tmpl.ID)
	assert.Equal(t, "Camry", tmpl.Model)
	assert.Equal(t, 1, tmpl.Version, "the version the worker stamped, not zero")
	require.Len(t, tmpl.Trims, 1)
	assert.Equal(t, "Base", tmpl.Trims[0].Name)
}

// The invariant the return value exists for: the first decode of a VIN and
// every decode after it answer the same thing. The created template, run
// through the matcher Handle runs, must resolve exactly as the stored template
// resolves for the cached path -- Base, exact, version 1 -- rather than the
// empty match block a nil template produced.
func TestTheCreatedTemplateResolvesTheWayTheNextDecodeWill(t *testing.T) {
	fw := &fakeWorker{}
	catalog := newFakeWorker(t, fw)

	created, err := createOrAdoptTemplate(context.Background(), catalog, "Toyota", firstDecode)
	require.NoError(t, err)
	require.NotNil(t, created)

	// What the next decode of the same VIN reads back out of the catalog.
	var stored coremodels.Template
	require.NoError(t, json.Unmarshal([]byte(fw.stored), &stored))

	sig := services.MatchSignals{VIN: "4T1C11AK8NU123456", StyleName: "LE"}
	fresh := services.MatchTrim(created, sig)
	cached := services.MatchTrim(&stored, sig)

	assert.Equal(t, cached, fresh, "the decode that created the definition must answer what the next one answers")
	assert.Equal(t, "Base", fresh.Trim)
	assert.Equal(t, services.MatchExact, fresh.Quality)
	assert.Equal(t, 1, fresh.TemplateVersion)
}

// Only a 412 means there is someone else's template to use. Any other refusal
// fails the decode.
func TestCreateOrAdoptTemplateFailsWhenTheWorkerRejectsTheTemplate(t *testing.T) {
	fw := &fakeWorker{rejectPut: http.StatusUnprocessableEntity}
	catalog := newFakeWorker(t, fw)

	tmpl, err := createOrAdoptTemplate(context.Background(), catalog, "Toyota", firstDecode)
	require.Error(t, err)
	assert.NotErrorIs(t, err, gateways.ErrTemplateExists)
	assert.Nil(t, tmpl)
	assert.Zero(t, fw.templateGets)
}

// The id is repaired and the body is not. definitions-worker checks that the
// body's model and the model segment of the id name the same thing, and it
// runs BOTH sides through the same character-class repair before comparing --
// so `up!` in the body matches segment `up`. dd-api's half of that contract is
// to keep sending the model as decoded: repairing the model to match the id
// would store "up" as the name of a car called "up!", and "Tribeca NY NJ" as
// the name of the Tribeca (NY/NJ). This pins the pair for both shapes the
// repair is known to produce, so neither side can move alone.
func TestCreateSendsTheDecodedModelAlongsideTheRepairedID(t *testing.T) {
	tests := []struct {
		make   string
		model  string
		year   int16
		wantID string
	}{
		{make: "Volkswagen", model: "up!", year: 2020, wantID: "volkswagen_up_2020"},
		{make: "Subaru", model: "Tribeca (NY/NJ)", year: 2020, wantID: "subaru_tribeca-ny-nj_2020"},
	}
	for _, tt := range tests {
		t.Run(tt.wantID, func(t *testing.T) {
			fw := &fakeWorker{}
			catalog := newFakeWorker(t, fw)

			// Exactly what Handle does with a decoded make/model/year.
			id, err := definitionIDForDecode(tt.make, tt.model, tt.year)
			require.NoError(t, err)
			require.Equal(t, tt.wantID, id, "the id the repair is there to produce")

			_, err = createOrAdoptTemplate(context.Background(), catalog, tt.make, coremodels.DeviceDefinitionTablelandModel{
				ID:         id,
				Model:      tt.model,
				Year:       int(tt.year),
				DeviceType: "vehicle",
			})
			require.NoError(t, err)

			require.Len(t, fw.putPaths, 1)
			assert.Equal(t, "/t/"+tt.wantID, fw.putPaths[0], "the write goes to the repaired id")
			assert.Equal(t, tt.wantID, gjson.Get(fw.stored, "id").String(), "the body's id is the repaired id")
			assert.Equal(t, tt.model, gjson.Get(fw.stored, "model").String(),
				"the body carries the model as decoded, never the id's repaired segment")
		})
	}
}

// The repair strips every character outside the worker's class, and never
// looks at what it produced. A model with no character in that class at all --
// a Japanese or Cyrillic name, which reaches a decode through vinInfoFromKnown
// whose source is not low-confidence -- leaves the id's middle segment empty,
// and ID_RE requires at least one character there. Creating it 422s, and so
// does every retry of every VIN of that model-year, forever. Refuse to build
// the id instead, the way the low-confidence gate refuses to name a template.
func TestDefinitionIDForDecodeRefusesAnIDNoTemplateCanExistAt(t *testing.T) {
	tests := []struct {
		make  string
		model string
		year  int16
	}{
		{make: "Toyota", model: "ハイエース", year: 2020},
		{make: "Lada", model: "Нива", year: 2021},
	}
	for _, tt := range tests {
		t.Run(tt.make+" "+tt.model, func(t *testing.T) {
			id, err := definitionIDForDecode(tt.make, tt.model, tt.year)
			require.Error(t, err, "an id with an empty segment must never reach a PUT")
			assert.Empty(t, id)

			var notFound *exceptions.NotFoundError
			require.ErrorAs(t, err, &notFound,
				"the decode surfaces this the way the low-confidence gate does: not found, so the client opens the manual picker")
			assert.ErrorIs(t, notFound.Err, common.ErrUnmintableDefinitionID)
		})
	}
}

// The gate must not fire for anything the worker accepts, repaired or not.
func TestDefinitionIDForDecodeBuildsEveryIDTheWorkerAccepts(t *testing.T) {
	tests := []struct {
		make   string
		model  string
		year   int16
		wantID string
	}{
		{make: "Toyota", model: "Camry", year: 2026, wantID: "toyota_camry_2026"},
		{make: "Volkswagen", model: "ID. Buzz", year: 2024, wantID: "volkswagen_id--buzz_2024"},
		{make: "Dodge", model: "Town & Country", year: 2012, wantID: "dodge_town-&-country_2012"},
		{make: "Kia", model: "Soul !EV!", year: 2020, wantID: "kia_soul-ev_2020"},
	}
	for _, tt := range tests {
		t.Run(tt.wantID, func(t *testing.T) {
			id, err := definitionIDForDecode(tt.make, tt.model, tt.year)
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

// The deployed configuration, end to end. DEFINITIONS_CATALOG_URL and
// DEFINITIONS_WORKER_URL are one hostname, so "read through the worker instead
// of the CDN" bypassed nothing: the read-back built a byte-identical request
// to the read the decode had already made, and Cloudflare answered it from
// cache. Against responses served `public, max-age=86400`, that means the
// read-back was served the very 404 it exists to see past, and the decode
// failed with "created by another writer but could not be read back" for a
// template that plainly exists.
func TestTheFreshReadIsNotServedTheCached404OnOneHost(t *testing.T) {
	fw := &fakeWorker{}
	catalog := newOneHostFakeWorker(t, fw)
	ctx := context.Background()

	// The decode's catalog read: no template yet. The edge now holds that 404
	// for a day.
	_, _, err := catalog.GetTemplateByID(ctx, firstDecode.ID)
	require.ErrorIs(t, err, gateways.ErrTemplateNotFound)

	// A curator saves the template through the worker a moment later.
	fw.stored = curatedTemplateJSON
	fw.version = 2

	// The cached read still answers 404. That is the edge doing its job, and
	// it is the reason the read-back cannot be an ordinary read.
	_, _, err = catalog.GetTemplateByID(ctx, firstDecode.ID)
	require.ErrorIs(t, err, gateways.ErrTemplateNotFound)
	require.Positive(t, fw.edgeHits, "the premise: the second identical read never reached the origin")

	tmpl, err := createOrAdoptTemplate(ctx, catalog, "Toyota", firstDecode)
	require.NoError(t, err, "the read-back must not be served the 404 the decode already saw")
	require.NotNil(t, tmpl)
	assert.Equal(t, 2, tmpl.Version)
	require.Len(t, tmpl.Trims, 2, "the curator's trims, read past the cache")
	assert.Equal(t, "Hybrid LE", tmpl.Trims[1].Name)
}

// What separates the two reads is the cache key, and nothing else. The path is
// identical because definitions-worker routes on url.pathname alone, so the
// query it never reads costs nothing there and is the whole difference at the
// edge. The cached read must stay cacheable -- it is the hot path -- so only
// the fresh read carries the parameter, and it must differ every time or the
// second fresh read of a day is served the first one's answer.
func TestOnlyTheFreshReadCarriesACacheBuster(t *testing.T) {
	fw := &fakeWorker{stored: curatedTemplateJSON}
	catalog := newFakeWorker(t, fw)
	ctx := context.Background()

	_, _, err := catalog.GetTemplateByIDFresh(ctx, firstDecode.ID)
	require.NoError(t, err)
	_, _, err = catalog.GetTemplateByIDFresh(ctx, firstDecode.ID)
	require.NoError(t, err)

	require.Len(t, fw.getQueries, 2)
	assert.NotEmpty(t, fw.getQueries[0], "a fresh read the edge has already seen is not a fresh read")
	assert.NotEqual(t, fw.getQueries[0], fw.getQueries[1], "two fresh reads must not share a cache key")
	for _, p := range fw.getPaths {
		assert.Equal(t, "/t/"+firstDecode.ID+".json", p, "the worker routes on the path; the query must not move it")
	}

	// The cached read is the hot path and must keep the CDN's cache key.
	fw.getQueries = nil
	_, _, err = catalog.GetTemplateByID(ctx, firstDecode.ID)
	require.ErrorIs(t, err, gateways.ErrTemplateNotFound, "the two-server fake's CDN always 404s")
	assert.Zero(t, fw.cdnRequests-1, "the cached read goes to the CDN")
	assert.Empty(t, fw.getQueries, "the cached read never reaches the worker")
}

// A catalog miss from one of these sources answers not found instead of
// creating a template (#312). Pinning the set keeps a new provider from being
// trusted, or an untrusted one from being dropped, without a decision.
func TestIsLowConfidenceSourceGatesOnlyTheUntrustedDecoders(t *testing.T) {
	for _, src := range []coremodels.DecodeProviderEnum{coremodels.Japan17VIN, coremodels.CarVXVIN, coremodels.AutoIsoProvider, coremodels.ElevaKaufmannProvider} {
		assert.True(t, isLowConfidenceSource(src), "%s", src)
	}
	for _, src := range []coremodels.DecodeProviderEnum{"drivly", "vincario", "datgroup", "tesla", ""} {
		assert.False(t, isLowConfidenceSource(src), "%q", src)
	}
}
