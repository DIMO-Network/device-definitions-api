package queries

import (
	"context"
	"math/big"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// UpsertDecoding's write needs Postgres, so the manufacturer resolution it
// depends on is tested on its own.

// A template's manufacturer.tokenId is optional, and GetTemplateByID reports an
// absent one as nil. Looking up manufacturer 0 in its place failed the upsert
// with an untyped error and never wrote the vin_numbers row.
func TestManufacturerNameForTemplateWithoutATokenIDUsesTheTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	catalog.EXPECT().GetManufacturerNameByID(gomock.Any(), gomock.Any()).Times(0)

	tmpl := &coremodels.Template{
		ID:           "toyota_camry_2026",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota"},
	}
	name, err := manufacturerNameForTemplate(context.Background(), catalog, tmpl, nil)
	require.NoError(t, err)
	assert.Equal(t, "Toyota", name)
}

func TestManufacturerNameForTemplateWithATokenIDAsksIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	catalog.EXPECT().GetManufacturerNameByID(gomock.Any(), big.NewInt(131)).Return("Toyota", nil)

	tmpl := &coremodels.Template{
		ID:           "toyota_camry_2020",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "TOYOTA", TokenID: 131},
	}
	name, err := manufacturerNameForTemplate(context.Background(), catalog, tmpl, big.NewInt(131))
	require.NoError(t, err)
	assert.Equal(t, "Toyota", name, "a known token id keeps identity as the source of the name")
}

// The fiber and gRPC error mappers type-assert rather than unwrap, so the
// not-found must come back bare to be answered as one.
func TestManufacturerNameForTemplateWithNeitherIsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	catalog.EXPECT().GetManufacturerNameByID(gomock.Any(), gomock.Any()).Times(0)

	_, err := manufacturerNameForTemplate(context.Background(), catalog, &coremodels.Template{ID: "toyota_camry_2026"}, nil)
	require.Error(t, err)
	assert.IsType(t, &exceptions.NotFoundError{}, err)
}
