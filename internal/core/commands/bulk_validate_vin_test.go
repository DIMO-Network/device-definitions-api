package commands

import (
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GetDeviceDefinitionByIDQuery answers with the catalog's *coremodels.Template
// since the R2 migration. Building a result row out of it needs no database.

func decodeResponse() *p_grpc.DecodeVinResponse {
	return &p_grpc.DecodeVinResponse{
		DefinitionId: "toyota_camry_2026",
		Year:         2026,
		Model:        "Camry",
	}
}

// The handler answers a template, and this is the type assertion that used to
// name a different one: a panic on the success path, for every VIN that
// decodes and whose definition is in the catalog.
func TestDecodedVINFromReadsTheTemplateTheHandlerAnswers(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:           "toyota_camry_2026",
		DeviceType:   "vehicle",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota", TokenID: 131},
		Model:        "Camry",
		Year:         2026,
	}

	got, err := decodedVINFrom("4T1C11AK8NU123456", decodeResponse(), tmpl)

	require.NoError(t, err)
	assert.Equal(t, "4T1C11AK8NU123456", got.VIN)
	assert.Equal(t, "toyota_camry_2026", got.DefinitionID)
	assert.Equal(t, int32(2026), got.DeviceYear)
	assert.Equal(t, "Camry", got.DeviceModel)
	assert.Equal(t, coremodels.Manufacturer{TokenID: 131, Name: "Toyota"}, got.DeviceMake)
}

// A template with no manufacturer token id is still a usable row: tokenId is
// optional in the contract, and the name is required.
func TestDecodedVINFromAcceptsATemplateWithNoManufacturerTokenID(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:           "toyota_camry_2026",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota"},
		Model:        "Camry",
		Year:         2026,
	}

	got, err := decodedVINFrom("4T1C11AK8NU123456", decodeResponse(), tmpl)

	require.NoError(t, err)
	assert.Equal(t, coremodels.Manufacturer{TokenID: 0, Name: "Toyota"}, got.DeviceMake)
	assert.Equal(t, "Camry", got.DeviceModel)
}

// If the handler's return type ever changes again, say so instead of taking
// down the request.
func TestDecodedVINFromReportsASurpriseTypeAsAnError(t *testing.T) {
	got, err := decodedVINFrom("4T1C11AK8NU123456", decodeResponse(), &coremodels.GetDeviceDefinitionQueryResult{})

	require.Error(t, err, "an unexpected result type is an error, not a panic")
	assert.Equal(t, DecodedVIN{}, got)

	_, err = decodedVINFrom("4T1C11AK8NU123456", decodeResponse(), nil)
	require.Error(t, err, "a nil result is an error, not a panic")
}
