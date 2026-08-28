package gateways

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	// ksuid and tableId were removed from the model deliberately; a template
	// carrying them would mean the worker regressed.
	var raw map[string]any
	require.NoError(t, json.Unmarshal([]byte(templateJSON), &raw))
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
}
