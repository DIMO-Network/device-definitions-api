package queries

import (
	"context"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/aarondl/null/v8"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// hydrateResponseFromVinNumber is the cached decode path, the one most decodes
// take. Its doc comment promises it answers what a fresh decode of the same VIN
// would, so these run it against a mocked catalog: no Postgres, because a
// vin_numbers row with no style id never touches the database.

const hydrateVIN = "4T1C11AK8NU123456"

// cachedVinNumber is a vin_numbers row as saveVinDecodeNumber writes one, with
// no style id so matchSignalsFromVinNumber makes no database call.
func cachedVinNumber() *models.VinNumber {
	return &models.VinNumber{
		Vin:              hydrateVIN,
		ManufacturerName: "Toyota",
		Year:             2026,
		DefinitionID:     "toyota_camry_2026",
		DecodeProvider:   null.StringFrom("drivly"),
	}
}

func hydrateTemplate() *coremodels.Template {
	return &coremodels.Template{
		ID:           "toyota_camry_2026",
		DeviceType:   "vehicle",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota", TokenID: 131},
		Model:        "Camry",
		Year:         2026,
		Attributes:   map[string]any{"number_of_doors": 4},
		Trims: []coremodels.Trim{
			{Name: "LE", Selectors: coremodels.TrimSelectors{StyleName: []string{"LE"}}, Attributes: map[string]any{"powertrain_type": "ICE"}},
			{Name: "Hybrid LE", Selectors: coremodels.TrimSelectors{StyleName: []string{"Hybrid LE"}}, Attributes: map[string]any{"powertrain_type": "HEV"}},
		},
		Version: 2,
	}
}

func hydrateHandler(t *testing.T) (DecodeVINQueryHandler, *mock_gateways.MockDeviceDefinitionCatalogService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	logger := zerolog.Nop()
	return DecodeVINQueryHandler{
		logger:                         &logger,
		deviceDefinitionCatalogService: catalog,
	}, catalog
}

// The same VIN answered from vin_numbers must carry the model a fresh decode
// answered, not an empty string that depends only on cache state.
func TestHydrateResponseFromVinNumberAnswersTheModel(t *testing.T) {
	dc, catalog := hydrateHandler(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), "toyota_camry_2026").Return(hydrateTemplate(), nil, nil)

	resp := dc.hydrateResponseFromVinNumber(context.Background(), cachedVinNumber())

	require.NotNil(t, resp)
	assert.Equal(t, "Camry", resp.Model, "the cached path must answer the model the fresh path does")
	// The rest of the fresh path's fields, so a future edit cannot drop one
	// the way Model was dropped.
	assert.Equal(t, "Toyota", resp.Manufacturer)
	assert.Equal(t, int32(2026), resp.Year)
	assert.Equal(t, "drivly", resp.Source)
	assert.Equal(t, "toyota_camry_2026", resp.DefinitionId)
	assert.Equal(t, int32(2), resp.TemplateVersion)
	assert.Equal(t, "model-only", resp.MatchQuality)
}
