package gateways

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
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
