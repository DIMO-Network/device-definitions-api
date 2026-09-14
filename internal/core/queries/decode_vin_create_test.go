package queries

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	templateGets int
	cdnRequests  int
}

func newFakeWorker(t *testing.T, fw *fakeWorker) gateways.DeviceDefinitionCatalogService {
	t.Helper()
	// The CDN still serves the 404 the decode's read saw.
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fw.cdnRequests++
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(cdn.Close)
	worker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			// identity-api's GraphQL endpoint.
			_, _ = w.Write([]byte(`{"data":{"manufacturer":{"tokenId":131,"name":"Toyota"}}}`))
		case strings.HasPrefix(r.URL.Path, "/schema/"):
			_, _ = w.Write([]byte(`{"id":"vehicle","attributes":[]}`))
		case r.Method == http.MethodPut:
			fw.puts = append(fw.puts, r.Header.Get("If-None-Match"))
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
			body, _ := io.ReadAll(r.Body)
			fw.stored = string(body)
			_, _ = w.Write(body)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/t/"):
			fw.templateGets++
			if fw.stored == "" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(fw.stored))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(worker.Close)

	identity, err := url.Parse(worker.URL)
	require.NoError(t, err)
	logger := zerolog.Nop()
	return gateways.NewDeviceDefinitionCatalogService(&config.Settings{
		DefinitionsCatalogURL:  cdn.URL,
		DefinitionsWorkerURL:   worker.URL,
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
	assert.Equal(t, 1, fw.templateGets, "the stored template is read back through the worker")
	assert.Zero(t, fw.cdnRequests, "the CDN may still be serving the 404")

	assert.Equal(t, 2, tmpl.Version)
	require.Len(t, tmpl.Trims, 2, "the curated trims, not a single Base trim")
	assert.Equal(t, "Hybrid LE", tmpl.Trims[1].Name)
}

// When this decode's create is the one that lands, the decode carries on as it
// always has: nothing to adopt, and no read-back.
func TestCreateOrAdoptTemplateReturnsNothingWhenItWroteTheTemplate(t *testing.T) {
	fw := &fakeWorker{}
	catalog := newFakeWorker(t, fw)

	tmpl, err := createOrAdoptTemplate(context.Background(), catalog, "Toyota", firstDecode)
	require.NoError(t, err)
	assert.Nil(t, tmpl)
	assert.Equal(t, []string{"*"}, fw.puts)
	assert.NotEmpty(t, fw.stored)
	assert.Zero(t, fw.templateGets)
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
