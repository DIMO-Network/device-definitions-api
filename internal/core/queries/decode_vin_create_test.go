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
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
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
