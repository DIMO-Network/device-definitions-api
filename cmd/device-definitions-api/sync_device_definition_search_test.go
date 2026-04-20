package main

import (
	"context"
	"errors"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
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
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 42, Name: "Toyota"}

	onChain.EXPECT().
		QueryDefinitionsCustom(gomock.Any(), 42, "", 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{
			defToyota(2006, "Camry", "id-old"),
			defToyota(2007, "Camry", "id-keep"),
			defToyota(2020, "Prius", "id-new"),
		}, nil)

	docs, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.Equal(t, "id-keep", docs[0].ID)
	assert.Equal(t, "id-new", docs[1].ID)
}

func TestBuildManufacturerDocuments_PopulatesFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 7, Name: "Land Rover"}

	onChain.EXPECT().
		QueryDefinitionsCustom(gomock.Any(), 7, "", 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{
			{ID: "ddid-1", Model: "Range Rover", Year: 2021, ImageURI: "https://img/rr"},
		}, nil)

	docs, err := buildManufacturerDocuments(context.Background(), onChain, dm)
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
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 1, Name: "Honda"}

	// 10 rows on page 0 is < 500 → loop should break without requesting page 1.
	page := make([]coremodels.DeviceDefinitionTablelandModel, 10)
	for i := range page {
		page[i] = defToyota(2020, "Civic", "id")
	}
	onChain.EXPECT().
		QueryDefinitionsCustom(gomock.Any(), 1, "", 0).
		Return(page, nil).
		Times(1)

	docs, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	assert.Len(t, docs, 10)
}

func TestBuildManufacturerDocuments_PagesUntilShortPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 1, Name: "Honda"}

	full := make([]coremodels.DeviceDefinitionTablelandModel, tablelandPageSize)
	for i := range full {
		full[i] = defToyota(2020, "Civic", "id")
	}
	short := []coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Accord", "id-last")}

	gomock.InOrder(
		onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 1, "", 0).Return(full, nil),
		onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 1, "", 1).Return(full, nil),
		onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 1, "", 2).Return(short, nil),
	)

	docs, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.NoError(t, err)
	assert.Len(t, docs, 2*tablelandPageSize+1)
}

func TestBuildManufacturerDocuments_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	dm := coremodels.Manufacturer{TokenID: 1, Name: "Honda"}

	boom := errors.New("tableland down")
	onChain.EXPECT().
		QueryDefinitionsCustom(gomock.Any(), 1, "", 0).
		Return(nil, boom)

	_, err := buildManufacturerDocuments(context.Background(), onChain, dm)
	require.ErrorIs(t, err, boom)
}

func TestRunSearchSync_FlushesPerManufacturer(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
		{TokenID: 2, Name: "Toyota"},
	}, nil)

	onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 1, "", 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)
	onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 2, "", 0).
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

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search")
	require.NoError(t, err)
}

func TestRunSearchSync_SkipsMakeWithNoEligibleDefs(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 9, Name: "Studebaker"},
	}, nil)

	// All pre-2007 → filtered out → builder returns zero docs.
	onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 9, "", 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(1950, "Champion", "s1")}, nil)

	// No UpsertDocuments expectation → gomock will fail the test if it's called.

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search")
	require.NoError(t, err)
}

func TestRunSearchSync_PropagatesManufacturersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	boom := errors.New("identity down")
	identity.EXPECT().GetManufacturers().Return(nil, boom)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search")
	require.ErrorIs(t, err, boom)
}

func TestRunSearchSync_PropagatesUpsertError(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
	}, nil)
	onChain.EXPECT().QueryDefinitionsCustom(gomock.Any(), 1, "", 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)

	boom := errors.New("typesense down")
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).Return(boom)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search")
	require.ErrorIs(t, err, boom)
}
