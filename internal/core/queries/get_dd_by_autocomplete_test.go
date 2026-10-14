package queries

import (
	"context"
	"testing"

	mock_search "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Autocomplete groups on definition_id, so Typesense answers in grouped_hits
// with hits unset. The handler used to dereference result.Hits unconditionally,
// which panics on exactly this shape, and before grouping it listed one entry
// per trim -- ten identical "Toyota Camry" suggestions for one Camry.
func TestGetAllDeviceDefinitionByAutocompleteQuery_GroupedHits(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mock_search.NewMockTypesenseAPIService(ctrl)
	svc.EXPECT().Autocomplete(gomock.Any(), "camry").
		Return(searchResultFromJSON(t, workerGroupedSearchJSON), nil)

	handler := NewGetAllDeviceDefinitionByAutocompleteQueryHandler(svc)
	out, err := handler.Handle(context.Background(), &GetAllDeviceDefinitionByAutocompleteQuery{Query: "camry"})
	require.NoError(t, err)
	result, ok := out.(*GetAllDeviceDefinitionByAutocompleteQueryResult)
	require.True(t, ok)

	// Two definitions, not the three trim documents behind them.
	require.Len(t, result.Items, 2)

	// The id is the definition slug, not the per-trim document id
	// ("toyota_camry_2020#LE"), and the label is the definition's name, not the
	// leading trim's ("2020 Toyota Camry LE").
	assert.Equal(t, "toyota_camry_2020", result.Items[0].ID)
	assert.Equal(t, "2020 Toyota Camry", result.Items[0].Name)
	assert.Equal(t, "ford_f-150_2021", result.Items[1].ID)
	assert.Equal(t, "2021 Ford F-150", result.Items[1].Name)
}

// Ungrouped hits remain the fallback, and a document missing fields must decode
// to zero values rather than panic on a type assertion.
func TestGetAllDeviceDefinitionByAutocompleteQuery_UngroupedAndMissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mock_search.NewMockTypesenseAPIService(ctrl)
	svc.EXPECT().Autocomplete(gomock.Any(), "up").
		Return(searchResultFromJSON(t, `{
  "found": 2,
  "hits": [
    {"document": {"id": "volkswagen_up_2019#Base", "definition_id": "volkswagen_up_2019",
                  "name": "2019 Volkswagen up", "make": "Volkswagen", "model": "up"}},
    {"document": {"id": "orphan#Base", "name": "no definition id here"}}
  ]
}`), nil)

	handler := NewGetAllDeviceDefinitionByAutocompleteQueryHandler(svc)
	var out interface{}
	var err error
	require.NotPanics(t, func() {
		out, err = handler.Handle(context.Background(), &GetAllDeviceDefinitionByAutocompleteQuery{Query: "up"})
	})
	require.NoError(t, err)
	result := out.(*GetAllDeviceDefinitionByAutocompleteQueryResult)

	// The document with no definition_id is dropped: an autocomplete entry whose
	// id cannot be looked up is not a usable suggestion.
	require.Len(t, result.Items, 1)
	assert.Equal(t, "volkswagen_up_2019", result.Items[0].ID)
	// year is absent, so definitionName falls back to the document's own name.
	assert.Equal(t, "2019 Volkswagen up", result.Items[0].Name)
}

// An empty result set must produce an empty list, not null, and must not panic
// on the absent hits/grouped_hits fields.
func TestGetAllDeviceDefinitionByAutocompleteQuery_NoHits(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mock_search.NewMockTypesenseAPIService(ctrl)
	svc.EXPECT().Autocomplete(gomock.Any(), "zzzz").
		Return(searchResultFromJSON(t, `{"found": 0}`), nil)

	handler := NewGetAllDeviceDefinitionByAutocompleteQueryHandler(svc)
	var out interface{}
	var err error
	require.NotPanics(t, func() {
		out, err = handler.Handle(context.Background(), &GetAllDeviceDefinitionByAutocompleteQuery{Query: "zzzz"})
	})
	require.NoError(t, err)
	result := out.(*GetAllDeviceDefinitionByAutocompleteQueryResult)
	assert.NotNil(t, result.Items)
	assert.Empty(t, result.Items)
}
