package gateways

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// manufacturer.tokenId is optional in the worker's schema (minimum 1 when
// present). Reporting an absent one as token id 0 sent UpsertDecoding to look
// up manufacturer 0, and made addvin reject a template that exists because 0
// never equals the caller's manufacturer.
const templateWithoutTokenIDJSON = `{
  "id": "toyota_camry_2026",
  "deviceType": "vehicle",
  "manufacturer": {"slug": "toyota", "name": "Toyota"},
  "model": "Camry",
  "year": 2026,
  "attributes": {"number_of_doors": 4},
  "trims": [{"name": "Base", "attributes": {}}],
  "version": 1
}`

func TestGetTemplateByIDReportsAnAbsentTokenIDAsUnknown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(templateWithoutTokenIDJSON))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: srv.URL}, &logger)

	tmpl, tokenID, err := svc.GetTemplateByID(context.Background(), "toyota_camry_2026")
	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.Nil(t, tokenID, "an absent tokenId is unknown, not manufacturer 0")
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
		if r.Method == http.MethodPost {
			// identity-api's GraphQL endpoint; createSvc points identity here.
			_, _ = w.Write([]byte(identityToyotaJSON))
			return
		}
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
	// identity points at the stub too: Create resolves the manufacturer there
	// before it writes anything.
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

// What identity-api answers for a minted make.
const identityToyotaJSON = `{"data":{"manufacturer":{"tokenId":131,"name":"Toyota","tableId":0,"owner":"0x0000000000000000000000000000000000000000"}}}`

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
	// tokenId is optional in the schema, but a create must never omit it:
	// UpsertDecoding and addvin both resolve the template through it.
	assert.Equal(t, map[string]any{"slug": "toyota", "name": "Toyota", "tokenId": float64(131)}, (*body)["manufacturer"])
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

// The worker's PUT is an upsert unless told otherwise. Create used to read the
// template and then write unconditionally, and in between a Console curator
// could save the same id with its trims: the create replaced them with a single
// Base trim and reported success. The existence check is now the write's own
// precondition, so there is no read before it to race.
func TestCreateIsCreateOnly(t *testing.T) {
	paths := []string{}
	var ifNoneMatch []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost:
			_, _ = w.Write([]byte(identityToyotaJSON))
		case strings.HasPrefix(r.URL.Path, "/schema/"):
			_, _ = w.Write([]byte(vocabularyJSON))
		case r.Method == http.MethodPut:
			ifNoneMatch = append(ifNoneMatch, r.Header.Get("If-None-Match"))
			_, _ = w.Write([]byte(`{"id":"toyota_camry_2020"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.Equal(t, []string{"*"}, ifNoneMatch, "exactly one PUT, and it must be create-only")
	for _, p := range paths {
		assert.False(t, strings.HasPrefix(p, "GET /t/"), "a read before the write is the race, not a guard against it: %s", p)
	}
}

// A 412 to If-None-Match: * means a template is already stored at the id. It is
// the one refusal a caller may treat as someone else's success, so it is the
// typed sentinel, and the write is neither retried nor forced.
func TestCreateReportsAnExistingTemplateAsErrTemplateExists(t *testing.T) {
	puts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			_, _ = w.Write([]byte(identityToyotaJSON))
		case strings.HasPrefix(r.URL.Path, "/schema/"):
			_, _ = w.Write([]byte(vocabularyJSON))
		case r.Method == http.MethodPut:
			puts++
			w.WriteHeader(http.StatusPreconditionFailed)
			_, _ = w.Write([]byte(`{"error":"template toyota_camry_2020 already exists","expected":null,"actual":2}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTemplateExists)
	assert.Nil(t, id)
	assert.Equal(t, 1, puts, "the write must not be retried or forced")
}

// Every other refusal is a failure. An invalid body must not be mistaken for a
// template someone else stored, and neither may an outage.
func TestCreateSurfacesOtherWorkerFailures(t *testing.T) {
	for _, status := range []int{http.StatusUnprocessableEntity, http.StatusInternalServerError} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost:
					_, _ = w.Write([]byte(identityToyotaJSON))
				case strings.HasPrefix(r.URL.Path, "/schema/"):
					_, _ = w.Write([]byte(vocabularyJSON))
				case r.Method == http.MethodPut:
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"errors":["rejected"]}`))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
			require.Error(t, err)
			assert.NotErrorIs(t, err, ErrTemplateExists)
			assert.Nil(t, id)
			assert.Contains(t, err.Error(), strconv.Itoa(status))
		})
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
		if r.Method == http.MethodPost {
			// identity resolves the make, so the vocabulary is what fails.
			_, _ = w.Write([]byte(identityToyotaJSON))
			return
		}
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

// The on-chain create this replaced refused a make identity could not resolve.
// Writing the template without a tokenId instead produced definitions that
// UpsertDecoding and addvin could not use, so nothing may be written at all.
func TestCreateRefusesAManufacturerIdentityCannotResolve(t *testing.T) {
	for name, identityAnswer := range map[string]string{
		"not minted":  `{"data":{"manufacturer":null}}`,
		"no token id": `{"data":{"manufacturer":{"tokenId":0,"name":"Toyota"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			puts := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost:
					_, _ = w.Write([]byte(identityAnswer))
				case r.Method == http.MethodPut:
					puts++
					w.WriteHeader(http.StatusOK)
				case strings.HasPrefix(r.URL.Path, "/schema/"):
					_, _ = w.Write([]byte(vocabularyJSON))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			id, err := createSvc(srv).Create(context.Background(), "Toyota", newDefinition)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrManufacturerUnresolved)
			assert.Nil(t, id)
			assert.Zero(t, puts, "nothing may be written for a manufacturer identity cannot resolve")
		})
	}
}

// The legacy flat shape has exactly one slot per attribute. Filling it from an
// arbitrary trim is what produced the record claiming powertrain ICE while
// carrying a hybrid's tank size, so only attributes shared by every trim cross
// over. The Camry fixture holds ICE and HEV trims with different tank sizes.
func TestGetDeviceDefinitionByIDCarriesOnlySharedAttributes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(templateJSON))
	}))
	defer srv.Close()

	dd, err := createSvc(srv).GetDeviceDefinitionByID(context.Background(), nil, "toyota_camry_2020")
	require.NoError(t, err)
	require.NotNil(t, dd)
	require.NotNil(t, dd.Metadata)

	got := map[string]string{}
	for _, a := range dd.Metadata.DeviceAttributes {
		got[a.Name] = a.Value
	}
	// Shared by every trim, and rendered without exponent notation.
	assert.Equal(t, "4", got["number_of_doors"])
	assert.Equal(t, "sedan", got["vehicle_type"])
	// Trim-varying: present on the trims, absent here rather than guessed.
	assert.NotContains(t, got, "powertrain_type")
	assert.NotContains(t, got, "fuel_tank_capacity_gal")
	assert.NotContains(t, got, "mpg_city")

	// ksuid is gone from the contract; inventing one would put a value in a
	// field consumers read as an identifier.
	assert.Empty(t, dd.KSUID)
}

func TestGetDeviceDefinitionByIDIsNilForAnotherManufacturer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(templateJSON))
	}))
	defer srv.Close()

	dd, err := createSvc(srv).GetDeviceDefinitionByID(context.Background(), big.NewInt(13), "toyota_camry_2020")
	require.NoError(t, err)
	assert.Nil(t, dd, "a template owned by another manufacturer must not resolve")
}

// Ownership can only be compared when both sides are known. A template with no
// tokenId belongs to no token, so it cannot belong to another manufacturer's.
func TestGetDeviceDefinitionByIDResolvesATemplateWithoutATokenID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(templateWithoutTokenIDJSON))
	}))
	defer srv.Close()

	dd, err := createSvc(srv).GetDeviceDefinitionByID(context.Background(), big.NewInt(131), "toyota_camry_2026")
	require.NoError(t, err)
	require.NotNil(t, dd, "a template that exists must resolve for the manufacturer whose id it carries")
	assert.Equal(t, "toyota_camry_2026", dd.ID)
}

func TestGetDeviceDefinitionByIDIsNilWhenMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dd, err := createSvc(srv).GetDeviceDefinitionByID(context.Background(), nil, "toyota_camry_2020")
	require.NoError(t, err, "not found is (nil, nil) here; callers branch on it and none create from it")
	assert.Nil(t, dd)
}

// Any other status is a catalog problem and must not be reclassified as
// not-found: that conflation is how an outage gets answered with a write.
func TestGetDeviceDefinitionByIDPropagatesCatalogFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dd, err := createSvc(srv).GetDeviceDefinitionByID(context.Background(), nil, "toyota_camry_2020")
	require.Error(t, err)
	assert.Nil(t, dd)
}

// GetTemplateByID reports an absent tokenId as nil. A lookup handed that, or a
// 0, must fail cleanly: not dereference nil, and not ask identity for a
// manufacturer that cannot exist.
func TestGetManufacturerNameByIDRejectsAnUnknownTokenID(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"data":{"manufacturers":{"totalCount":1,"nodes":[{"tokenId":131,"name":"Toyota"}]}}}`))
	}))
	defer srv.Close()

	svc := createSvc(srv)
	for _, tokenID := range []*big.Int{big.NewInt(0), nil} {
		name, err := svc.GetManufacturerNameByID(context.Background(), tokenID)
		require.Error(t, err, "token id %v", tokenID)
		assert.Empty(t, name)
	}
	assert.Zero(t, calls, "an unknown token id is not worth a round trip to identity")
}

// The definition path and the device-style path flatten the same template
// attributes, so they share one renderer and one order: a number is never
// served in scientific notation, and the array does not shuffle between calls.
func TestTemplateToDefinitionModelRendersAndOrdersAttributes(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:         "bugatti_veyron-16.4_2006",
		Model:      "Veyron 16.4",
		Year:       2006,
		DeviceType: "vehicle",
		Attributes: map[string]any{
			"base_msrp":              float64(1250000),
			"fuel_tank_capacity_gal": float64(15.8),
			"number_of_doors":        float64(2),
			"driven_wheels":          "AWD",
		},
	}

	want := []coremodels.DeviceTypeAttribute{
		{Name: "base_msrp", Value: "1250000"},
		{Name: "driven_wheels", Value: "AWD"},
		{Name: "fuel_tank_capacity_gal", Value: "15.8"},
		{Name: "number_of_doors", Value: "2"},
	}
	for i := 0; i < 20; i++ {
		assert.Equal(t, want, templateToDefinitionModel(tmpl).Metadata.DeviceAttributes)
	}
}
