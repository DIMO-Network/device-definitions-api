package main

import (
	"context"
	"errors"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func defToyota(year int, model, id string) coremodels.DeviceDefinitionTablelandModel {
	return coremodels.DeviceDefinitionTablelandModel{
		ID:       id,
		Model:    model,
		Year:     year,
		ImageURI: "https://img/" + id,
	}
}

func TestBuildManufacturerDocuments_FiltersPre2007(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 42, Name: "Toyota"}

	onChain.EXPECT().
		QueryDefinitionsByManufacturer(gomock.Any(), 42, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{
			defToyota(2006, "Camry", "id-old"),
			defToyota(2007, "Camry", "id-keep"),
			defToyota(2020, "Prius", "id-new"),
		}, nil)

	docs, _, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.Equal(t, "id-keep", docs[0].ID)
	assert.Equal(t, "id-new", docs[1].ID)
}

func TestBuildManufacturerDocuments_PopulatesFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 7, Name: "Land Rover"}

	onChain.EXPECT().
		QueryDefinitionsByManufacturer(gomock.Any(), 7, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{
			{ID: "ddid-1", Model: "Range Rover", Year: 2021, ImageURI: "https://img/rr"},
		}, nil)

	docs, _, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	d := docs[0]
	assert.Equal(t, "ddid-1", d.ID)
	assert.Equal(t, "ddid-1", d.DeviceDefinitionID)
	assert.Equal(t, "ddid-1", d.DefinitionID)
	assert.Equal(t, "Land Rover", d.Make)
	assert.Equal(t, "land-rover", d.MakeSlug)
	assert.Equal(t, 7, d.ManufacturerTokenID)
	assert.Equal(t, "Range Rover", d.Model)
	assert.Equal(t, "range-rover", d.ModelSlug)
	assert.Equal(t, 2021, d.Year)
	assert.Equal(t, "2021 Land Rover Range Rover", d.Name)
	assert.Equal(t, "https://img/rr", d.ImageURL)
	assert.Equal(t, searchDefaultScore, d.Score)
}

func TestBuildManufacturerDocuments_TerminatesOnShortPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 1, Name: "Honda"}

	// 10 rows on page 0 is < 500 → loop should break without requesting page 1.
	page := make([]coremodels.DeviceDefinitionTablelandModel, 10)
	for i := range page {
		page[i] = defToyota(2020, "Civic", "id")
	}
	onChain.EXPECT().
		QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return(page, nil).
		Times(1)

	docs, _, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	assert.Len(t, docs, 10)
}

func TestBuildManufacturerDocuments_PagesUntilShortPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 1, Name: "Honda"}

	full := make([]coremodels.DeviceDefinitionTablelandModel, gateways.CatalogPageSize)
	for i := range full {
		full[i] = defToyota(2020, "Civic", "id")
	}
	short := []coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Accord", "id-last")}

	gomock.InOrder(
		onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).Return(full, nil),
		onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 1).Return(full, nil),
		onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 2).Return(short, nil),
	)

	docs, _, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	assert.Len(t, docs, 2*gateways.CatalogPageSize+1)
}

func TestBuildManufacturerDocuments_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 1, Name: "Honda"}

	boom := errors.New("tableland down")
	onChain.EXPECT().
		QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return(nil, boom)

	_, _, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.ErrorIs(t, err, boom)
}

func TestRunSearchSync_FlushesPerManufacturer(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
		{TokenID: 2, Name: "Toyota"},
	}, nil)

	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 2, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Camry", "t1")}, nil)

	// One upsert per manufacturer, each with exactly that make's docs.
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, docs []SearchEntryItem) error {
			require.Len(t, docs, 1)
			assert.Equal(t, "Honda", docs[0].Make)
			return nil
		})
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, docs []SearchEntryItem) error {
			require.Len(t, docs, 1)
			assert.Equal(t, "Toyota", docs[0].Make)
			return nil
		})

	// Both makes are still in the catalog, so the prune pass deletes nothing.
	indexer.EXPECT().ExportIDs(gomock.Any(), "dd-search").Return([]string{"h1", "t1"}, nil)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.NoError(t, err)
}

func TestRunSearchSync_SkipsMakeWithNoEligibleDefs(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 9, Name: "Studebaker"},
	}, nil)

	// All pre-2007 → filtered out → builder returns zero docs.
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 9, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(1950, "Champion", "s1")}, nil)

	// No UpsertDocuments expectation → gomock will fail the test if it's called.
	// The definition still exists though, so the prune pass runs and keeps it.
	indexer.EXPECT().ExportIDs(gomock.Any(), "dd-search").Return([]string{"s1"}, nil)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.NoError(t, err)
}

func TestRunSearchSync_PropagatesManufacturersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	boom := errors.New("identity down")
	identity.EXPECT().GetManufacturers().Return(nil, boom)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.ErrorIs(t, err, boom)
}

func TestRunSearchSync_PropagatesUpsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
	}, nil)
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)

	boom := errors.New("typesense down")
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).Return(boom)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.ErrorIs(t, err, boom)
}

func TestRunSearchSync_PrunesDefinitionsMissingFromCatalog(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
	}, nil)
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).Return(nil)

	// The index still carries a definition the catalog no longer has.
	indexer.EXPECT().ExportIDs(gomock.Any(), "dd-search").
		Return([]string{"h1", "deleted_1"}, nil)
	indexer.EXPECT().DeleteDocuments(gomock.Any(), "dd-search", []string{"deleted_1"}).
		Return(nil)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.NoError(t, err)
}

func TestRunSearchSync_DoesNotPruneDefinitionsThatAreStillInTheCatalog(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
	}, nil)
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).Return(nil)
	indexer.EXPECT().ExportIDs(gomock.Any(), "dd-search").Return([]string{"h1"}, nil)
	// No DeleteDocuments expectation: gomock fails the test if it is called.

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.NoError(t, err)
}
