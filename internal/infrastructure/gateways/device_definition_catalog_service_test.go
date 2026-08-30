package gateways

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Worker-written document, verbatim shape.
const catalogDocJSON = `{
  "id": "dodge_town-&-country_2012",
  "ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
  "model": "Town & Country",
  "year": 2012,
  "devicetype": "vehicle",
  "imageuri": "https://image",
  "metadata": {"device_attributes": [{"name": "powertrain_type", "value": "ICE"}]},
  "manufacturer": {"tokenId": 22, "slug": "dodge", "name": "Dodge"},
  "createdAt": "2026-08-19T00:00:00.000Z",
  "updatedAt": "2026-08-19T00:00:00.000Z"
}`

// The embedded tableland model has a custom UnmarshalJSON; without the
// catalogDoc override it gets promoted and Manufacturer silently stays zero.
func TestCatalogDocUnmarshalKeepsManufacturer(t *testing.T) {
	var doc catalogDoc
	require.NoError(t, json.Unmarshal([]byte(catalogDocJSON), &doc))

	assert.Equal(t, "dodge_town-&-country_2012", doc.ID)
	assert.Equal(t, "Town & Country", doc.Model)
	assert.Equal(t, 2012, doc.Year)
	assert.Equal(t, "vehicle", doc.DeviceType)
	assert.Equal(t, "https://image", doc.ImageURI)
	require.NotNil(t, doc.Metadata)
	require.Len(t, doc.Metadata.DeviceAttributes, 1)
	assert.Equal(t, "powertrain_type", doc.Metadata.DeviceAttributes[0].Name)

	assert.Equal(t, 22, doc.Manufacturer.TokenID)
	assert.Equal(t, "dodge", doc.Manufacturer.Slug)
	assert.Equal(t, "Dodge", doc.Manufacturer.Name)
}

func TestCatalogDocUnmarshalToleratesEmptyMetadata(t *testing.T) {
	var doc catalogDoc
	require.NoError(t, json.Unmarshal([]byte(`{"id":"bmw_x5_2019","model":"X5","year":2019,"metadata":"","manufacturer":{"tokenId":13,"slug":"bmw","name":"BMW"}}`), &doc))
	assert.Nil(t, doc.Metadata)
	assert.Equal(t, 13, doc.Manufacturer.TokenID)
}

func TestCatalogManifestDecode(t *testing.T) {
	var m catalogManifest
	require.NoError(t, json.Unmarshal([]byte(`{"updatedAt":"2026-08-19T00:00:00.000Z","count":1,"definitions":[`+catalogDocJSON+`]}`), &m))
	require.Len(t, m.Definitions, 1)
	assert.Equal(t, 22, m.Definitions[0].Manufacturer.TokenID)
	assert.Equal(t, "dodge_town-&-country_2012", m.Definitions[0].ID)
}

// An unconfigured worker URL used to make every write a no-op that still
// reported success: Create returned the id, Delete logged "Deleted", and the
// API answered 200 while nothing reached R2. A missing write endpoint is a
// misconfiguration, not a mode of operation.
func TestWritesFailWhenWorkerURLUnset(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: "http://127.0.0.1:1"}, &logger)

	id, err := svc.Delete(context.Background(), "Toyota", "toyota_camry_2020")
	require.Error(t, err, "a delete with no worker configured must not report success")
	assert.Nil(t, id)
	assert.Contains(t, err.Error(), "not configured")
}

// Verbatim trim of the Camry template the definitions-worker pipeline
// produces.
const templateJSON = `{
  "id": "toyota_camry_2020",
  "deviceType": "vehicle",
  "manufacturer": {"slug": "toyota", "name": "Toyota", "tokenId": 131},
  "model": "Camry",
  "year": 2020,
  "attributes": {"number_of_doors": 4, "vehicle_type": "sedan"},
  "trims": [
    {"name": "LE", "selectors": {"manufacturerCode": ["2532"]},
     "attributes": {"powertrain_type": "ICE", "fuel_type": "gasoline", "fuel_tank_capacity_gal": 16, "mpg_city": 28}},
    {"name": "Hybrid LE", "selectors": {"manufacturerCode": ["2559"]},
     "attributes": {"powertrain_type": "HEV", "fuel_type": "gasoline", "fuel_tank_capacity_gal": 13.2, "mpg_city": 51}}
  ],
  "version": 3,
  "createdAt": "2026-08-27T00:00:00.000Z",
  "updatedAt": "2026-08-28T00:00:00.000Z"
}`

func TestTemplateUnmarshal(t *testing.T) {
	var tmpl coremodels.Template
	require.NoError(t, json.Unmarshal([]byte(templateJSON), &tmpl))

	assert.Equal(t, "toyota_camry_2020", tmpl.ID)
	assert.Equal(t, 2020, tmpl.Year)
	assert.Equal(t, 3, tmpl.Version)
	assert.Equal(t, "Toyota", tmpl.Manufacturer.Name)
	assert.Equal(t, 131, tmpl.Manufacturer.TokenID)

	// Attributes are typed, not stringified.
	assert.Equal(t, float64(4), tmpl.Attributes["number_of_doors"])

	require.Len(t, tmpl.Trims, 2)
	assert.Equal(t, "LE", tmpl.Trims[0].Name)
	assert.Equal(t, []string{"2532"}, tmpl.Trims[0].Selectors.ManufacturerCode)
	assert.Equal(t, "ICE", tmpl.Trims[0].Attributes["powertrain_type"])
	assert.Equal(t, 16.0, tmpl.Trims[0].Attributes["fuel_tank_capacity_gal"])
	assert.Equal(t, 13.2, tmpl.Trims[1].Attributes["fuel_tank_capacity_gal"])
}

func TestTemplateHasNoLegacyFields(t *testing.T) {
	// ksuid and tableId were removed from the model deliberately. Asserting
	// on the fixture JSON alone can never fail -- the fixture just doesn't
	// carry those keys, and nothing here reads the type. Round-tripping a
	// document that DOES carry them through coremodels.Template is what
	// would catch a regression: if the struct ever grew a field capturing
	// either key, it would come back out on re-marshal.
	withLegacyFields := `{
  "id": "toyota_camry_2020",
  "deviceType": "vehicle",
  "manufacturer": {"slug": "toyota", "name": "Toyota", "tokenId": 131},
  "model": "Camry",
  "year": 2020,
  "ksuid": "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
  "tableId": 42,
  "attributes": {},
  "trims": [{"name": "LE", "attributes": {}}],
  "version": 1
}`

	var tmpl coremodels.Template
	require.NoError(t, json.Unmarshal([]byte(withLegacyFields), &tmpl))

	out, err := json.Marshal(tmpl)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(out, &raw))
	assert.NotContains(t, raw, "ksuid")
	assert.NotContains(t, raw, "tableId")
}

func TestGetTemplateByIDReadsTheTemplateKey(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(templateJSON))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: srv.URL}, &logger)

	tmpl, tokenID, err := svc.GetTemplateByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	assert.Equal(t, "/t/toyota_camry_2020.json", gotPath)
	assert.Equal(t, "toyota_camry_2020", tmpl.ID)
	assert.Equal(t, int64(131), tokenID.Int64())
}

func TestGetTemplateByIDDoesNotFallBackToTheOldKey(t *testing.T) {
	// A 404 on t/<id>.json means the import has not run. Falling back to
	// definitions/<id>.json would serve the pre-migration flat record and hide
	// an incomplete import behind apparently-working decodes.
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: srv.URL}, &logger)

	_, _, err := svc.GetTemplateByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err)
	assert.Equal(t, []string{"/t/toyota_camry_2020.json"}, paths)
	// A genuine 404 must be the typed sentinel, checked by identity, so a
	// caller can tell "does not exist" apart from every other failure.
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}

// A 500 from the catalog is an outage, not a missing template. It must not
// satisfy errors.Is(err, ErrTemplateNotFound): a caller that reclassified it
// as not-found could react by creating a duplicate definition for a vehicle
// that already exists, during the worst possible moment to do so.
func TestGetTemplateByID500IsNotErrTemplateNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: srv.URL}, &logger)

	_, _, err := svc.GetTemplateByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrTemplateNotFound)
}

// GetTemplateByIDFresh shares fetchTemplateDocFrom with GetTemplateByID, so
// the no-fallback guarantee holds for it by inspection -- but this proves it
// for the worker-backed path itself rather than leaving it proven for only
// one of the two exported methods.
func TestGetTemplateByIDFreshDoesNotFallBackToTheOldKey(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{
		DefinitionsCatalogURL: srv.URL,
		DefinitionsWorkerURL:  srv.URL,
	}, &logger)

	_, _, err := svc.GetTemplateByIDFresh(context.Background(), "toyota_camry_2020")
	require.Error(t, err)
	assert.Equal(t, []string{"/t/toyota_camry_2020.json"}, paths)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}

// Create used to PUT the pre-migration /definitions/<id> route, which the new
// worker does not serve at all: every catalog miss on the decode path would
// have failed once deployed. These pin the route and the body shape, because
// both are only observable from outside the process.
func newCreateStub(t *testing.T, existing func(w http.ResponseWriter)) (*httptest.Server, *[]string, *map[string]any) {
	t.Helper()
	paths := []string{}
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/schema/") {
			_, _ = w.Write([]byte(vocabularyJSON))
			return
		}
		if r.Method == http.MethodGet {
			existing(w)
			return
		}
		if r.Method == http.MethodPut {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"toyota_camry_2020"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &paths, &body
}

func createSvc(srv *httptest.Server) DeviceDefinitionCatalogService {
	logger := zerolog.Nop()
	// identity points at the stub too: the manufacturer tokenId lookup is best
	// effort, and left unset it spends the http client's full timeout failing.
	identity, err := url.Parse(srv.URL)
	if err != nil {
		panic(err)
	}
	return NewDeviceDefinitionCatalogService(&config.Settings{
		DefinitionsCatalogURL:  srv.URL,
		DefinitionsWorkerURL:   srv.URL,
		DefinitionsWorkerToken: "test-token",
		IdentityAPIURL:         *identity,
	}, &logger)
}

// A trim of the real /schema/device-type-vehicle.json the worker serves.
const vocabularyJSON = `{
  "id": "vehicle",
  "attributes": [
    {"name": "fuel_type", "type": "enum", "options": ["gasoline", "diesel", "electric"]},
    {"name": "driven_wheels", "type": "enum", "options": ["FWD", "RWD", "AWD", "4WD"]},
    {"name": "number_of_doors", "type": "integer", "minimum": 1, "maximum": 8},
    {"name": "fuel_tank_capacity_gal", "type": "number", "minimum": 0, "maximum": 100}
  ]
}`

var newDefinition = coremodels.DeviceDefinitionTablelandModel{
	ID:         "toyota_camry_2020",
	Model:      "Camry",
	Year:       2020,
	DeviceType: "vehicle",
	KSUID:      "26G3iFH7Xc9Wvsw7pg6sD7uzoSS",
}

func TestCreateWritesATemplate(t *testing.T) {
	srv, paths, body := newCreateStub(t, func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) })

	id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
	require.NoError(t, err)
	require.NotNil(t, id)

	assert.Contains(t, *paths, "PUT /t/toyota_camry_2020", "the write must go to the template route")
	for _, p := range *paths {
		assert.NotContains(t, p, "/definitions/", "no write may touch the pre-migration route")
	}

	// The worker rejects a client-supplied server-owned field by name rather
	// than stripping it, so sending one fails the whole write.
	for _, k := range []string{"version", "createdAt", "updatedAt", "author"} {
		assert.NotContains(t, *body, k, "%s is server-owned and must not be sent", k)
	}
	assert.Equal(t, "toyota_camry_2020", (*body)["id"])
	assert.Equal(t, "vehicle", (*body)["deviceType"])
	assert.Equal(t, "Camry", (*body)["model"])
	assert.Equal(t, float64(2020), (*body)["year"])
	assert.Equal(t, map[string]any{"slug": "toyota", "name": "Toyota"}, (*body)["manufacturer"])
	// ksuid is gone from the contract; it must not ride along in any form.
	assert.NotContains(t, *body, "ksuid")

	// trims has minItems 1 in the schema: a model-year sold in one
	// configuration still has a trim.
	trims, ok := (*body)["trims"].([]any)
	require.True(t, ok, "trims must be present")
	require.Len(t, trims, 1)
	trim := trims[0].(map[string]any)
	assert.Equal(t, "Base", trim["name"])
	// A single-trim template has nothing to select on, and an empty selector
	// object is what the worker rejects as degenerate on a multi-trim template.
	assert.NotContains(t, trim, "selectors")
}

func TestCreateRefusesWhenTheTemplateExists(t *testing.T) {
	srv, paths, _ := newCreateStub(t, func(w http.ResponseWriter) {
		_, _ = w.Write([]byte(templateJSON))
	})

	id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
	require.Error(t, err, "create must not overwrite a curated template")
	assert.Nil(t, id)
	assert.Contains(t, err.Error(), "already exists")
	for _, p := range *paths {
		assert.NotEqual(t, "PUT /t/toyota_camry_2020", p, "nothing may be written")
	}
}

// The reason Create reads through the worker and checks the sentinel: a
// catalog outage returns 500, and treating that as "does not exist" is how a
// duplicate definition gets written mid-incident.
func TestCreateAbortsWhenTheCatalogIsDown(t *testing.T) {
	srv, paths, _ := newCreateStub(t, func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
	require.Error(t, err, "a 500 must abort the create, not authorise it")
	assert.Nil(t, id)
	for _, p := range *paths {
		assert.NotEqual(t, "PUT /t/toyota_camry_2020", p, "nothing may be written during an outage")
	}
}

func TestDeleteUsesTheTemplateRoute(t *testing.T) {
	srv, paths, _ := newCreateStub(t, func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) })

	id, err := createSvc(srv).Delete(context.Background(), "Toyota", "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Contains(t, *paths, "DELETE /t/toyota_camry_2020")
}

// The decode path's metadata is stringified under source-specific names. The
// contract requires typed values drawn from the DeviceType vocabulary, so the
// create folds them the same way the extraction pipeline does.
func TestCreateFoldsMetadataOntoTheVocabulary(t *testing.T) {
	srv, _, body := newCreateStub(t, func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) })

	withMetadata := newDefinition
	withMetadata.Metadata = &coremodels.DeviceDefinitionMetadata{
		DeviceAttributes: []coremodels.DeviceTypeAttribute{
			{Name: "fuel_type", Value: "Petrol"},
			{Name: "number_of_doors", Value: "4"},
			{Name: "fuel_tank_capacity_gal", Value: "15.800000"},
		},
	}
	_, err := createSvc(srv).Create(context.Background(), "Toyota", withMetadata)
	require.NoError(t, err)

	attrs, ok := (*body)["attributes"].(map[string]any)
	require.True(t, ok)
	// Folded to the vocabulary's spelling, and typed: 15.8, not "15.800000".
	assert.Equal(t, "gasoline", attrs["fuel_type"])
	assert.Equal(t, float64(4), attrs["number_of_doors"])
	assert.Equal(t, 15.8, attrs["fuel_tank_capacity_gal"])

	// Shared by every trim, so they belong on the template. The contract
	// forbids an attribute being in both places.
	trim := (*body)["trims"].([]any)[0].(map[string]any)
	assert.Empty(t, trim["attributes"])
}

func TestCreateDropsValuesItCannotMap(t *testing.T) {
	srv, _, body := newCreateStub(t, func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) })

	withJunk := newDefinition
	withJunk.Metadata = &coremodels.DeviceDefinitionMetadata{
		DeviceAttributes: []coremodels.DeviceTypeAttribute{
			{Name: "fuel_type", Value: "gasoline"},
			// Names a driven wheel count, not an axle: deliberately unmapped.
			{Name: "driven_wheels", Value: "4x2"},
			// Not in the vocabulary at all.
			{Name: "generation", Value: "6"},
			// The placeholders the contract makes unrepresentable.
			{Name: "number_of_doors", Value: "<nil>"},
		},
	}
	_, err := createSvc(srv).Create(context.Background(), "Toyota", withJunk)
	require.NoError(t, err)

	attrs := (*body)["attributes"].(map[string]any)
	assert.Equal(t, map[string]any{"fuel_type": "gasoline"}, attrs,
		"only values the vocabulary accepts may be written")
}

// Writing fewer attributes because a fetch failed is indistinguishable
// afterwards from the source never having carried them.
func TestCreateAbortsWhenTheVocabularyIsUnavailable(t *testing.T) {
	paths := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/schema/") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	withMetadata := newDefinition
	withMetadata.Metadata = &coremodels.DeviceDefinitionMetadata{
		DeviceAttributes: []coremodels.DeviceTypeAttribute{{Name: "fuel_type", Value: "Petrol"}},
	}
	id, err := createSvc(srv).Create(context.Background(), "Toyota", withMetadata)
	require.Error(t, err, "a create must not silently write an attribute-free template")
	assert.Nil(t, id)
	for _, p := range paths {
		assert.NotEqual(t, "PUT /t/toyota_camry_2020", p)
	}
}
