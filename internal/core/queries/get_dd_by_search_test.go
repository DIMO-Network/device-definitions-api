package queries

import (
	"context"
	"encoding/json"
	"testing"

	mock_search "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/typesense/typesense-go/typesense/api"
	"go.uber.org/mock/gomock"
)

// searchResultFromJSON builds the typesense-go result the way the client
// does: by decoding the wire shape. FacetCounts.Counts is an anonymous struct
// slice, so a literal is unwieldy and a real response is the more honest
// fixture anyway.
func searchResultFromJSON(t *testing.T, body string) *api.SearchResult {
	t.Helper()
	var result api.SearchResult
	require.NoError(t, json.Unmarshal([]byte(body), &result))
	return &result
}

// The definitions-worker indexes ONE DOCUMENT PER TRIM with the fields below.
// There is no device_definition_id (legacy ksuids are gone) and image_url may
// be empty, and each document's name carries its trim. A grouped search on
// definition_id returns each definition once, with its top trim document as the
// group's first hit.
const workerGroupedSearchJSON = `{
  "found": 2,
  "found_docs": 3,
  "out_of": 1000,
  "page": 1,
  "facet_counts": [
    {"field_name": "make", "counts": [{"count": 2, "value": "Toyota"}, {"count": 1, "value": "Ford"}]},
    {"field_name": "model", "counts": [{"count": 2, "value": "Camry"}, {"count": 1, "value": "F-150"}]},
    {"field_name": "year", "counts": [{"count": 2, "value": "2020"}, {"count": 1, "value": "2021"}]}
  ],
  "grouped_hits": [
    {
      "group_key": ["toyota_camry_2020"],
      "found": 2,
      "hits": [
        {"document": {"id": "toyota_camry_2020#LE", "definition_id": "toyota_camry_2020", "trim": "LE",
                      "name": "2020 Toyota Camry LE", "make": "Toyota", "make_slug": "toyota", "make_token_id": 131,
                      "model": "Camry", "model_slug": "camry", "year": 2020, "image_url": "", "score": 10}},
        {"document": {"id": "toyota_camry_2020#Hybrid LE", "definition_id": "toyota_camry_2020", "trim": "Hybrid LE",
                      "name": "2020 Toyota Camry Hybrid LE", "make": "Toyota", "make_slug": "toyota", "make_token_id": 131,
                      "model": "Camry", "model_slug": "camry", "year": 2020, "image_url": "", "score": 9}}
      ]
    },
    {
      "group_key": ["ford_f-150_2021"],
      "found": 1,
      "hits": [
        {"document": {"id": "ford_f-150_2021#XLT", "definition_id": "ford_f-150_2021", "trim": "XLT",
                      "name": "2021 Ford F-150 XLT", "make": "Ford", "make_slug": "ford", "make_token_id": 46,
                      "model": "F-150", "model_slug": "f-150", "year": 2021, "image_url": "https://img/f150.png", "score": 8}}
      ]
    }
  ]
}`

func TestGetAllDeviceDefinitionBySearchQuery_WorkerShapedGroupedHits(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mock_search.NewMockTypesenseAPIService(ctrl)
	svc.EXPECT().GetDeviceDefinitions(gomock.Any(), "camry", "", "", 0, 1, 20).
		Return(searchResultFromJSON(t, workerGroupedSearchJSON), nil)

	handler := NewGetAllDeviceDefinitionBySearchQueryHandler(svc)
	out, err := handler.Handle(context.Background(), &GetAllDeviceDefinitionBySearchQuery{Query: "camry", Page: 1, PageSize: 20})
	require.NoError(t, err)
	result, ok := out.(*GetAllDeviceDefinitionBySearchQueryResult)
	require.True(t, ok)

	// One item per definition, not per trim: the Camry group carries two
	// trims and must come back once.
	require.Len(t, result.DeviceDefinitions, 2)

	camry := result.DeviceDefinitions[0]
	assert.Equal(t, "toyota_camry_2020", camry.ID)
	// The legacy field stays on the wire for compatibility and carries the
	// same slug id; there is no other identifier to put in it.
	assert.Equal(t, camry.ID, camry.DeviceDefinitionID)
	assert.Equal(t, "2020 Toyota Camry", camry.Name)
	assert.Equal(t, "Toyota", camry.Make)
	assert.Equal(t, "Camry", camry.Model)
	assert.Equal(t, 2020, camry.Year)
	assert.Equal(t, "", camry.ImageURL)

	f150 := result.DeviceDefinitions[1]
	assert.Equal(t, "ford_f-150_2021", f150.ID)
	assert.Equal(t, "ford_f-150_2021", f150.DeviceDefinitionID)
	assert.Equal(t, 2021, f150.Year)
	assert.Equal(t, "https://img/f150.png", f150.ImageURL)

	// Facet parsing is unchanged.
	require.Len(t, result.Facets.Makes, 2)
	assert.Equal(t, GetAllDeviceDefinitionFacetItem{Name: "Toyota", Count: 2}, result.Facets.Makes[0])
	require.Len(t, result.Facets.Models, 2)
	assert.Equal(t, GetAllDeviceDefinitionFacetItem{Name: "Camry", Count: 2}, result.Facets.Models[0])
	require.Len(t, result.Facets.Years, 2)
	assert.Equal(t, GetAllDeviceDefinitionFacetItem{Name: "2020", Count: 2}, result.Facets.Years[0])

	assert.Equal(t, 2, result.Pagination.TotalItems)
	assert.Equal(t, 1, result.Pagination.TotalPages)
}

// A document with fields missing entirely -- no image_url, no year, and of
// course no device_definition_id -- must decode to zero values, never panic.
// Ungrouped hits are the fallback when the fake (or an older index) returns
// no grouped_hits.
func TestGetAllDeviceDefinitionBySearchQuery_MissingFieldsDoNotPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mock_search.NewMockTypesenseAPIService(ctrl)
	svc.EXPECT().GetDeviceDefinitions(gomock.Any(), "up", "", "", 0, 1, 10).
		Return(searchResultFromJSON(t, `{
  "found": 1,
  "facet_counts": [],
  "hits": [
    {"document": {"id": "volkswagen_up_2019#Base", "definition_id": "volkswagen_up_2019",
                  "name": "2019 Volkswagen up", "make": "Volkswagen", "model": "up"}}
  ]
}`), nil)

	handler := NewGetAllDeviceDefinitionBySearchQueryHandler(svc)
	var out interface{}
	var err error
	require.NotPanics(t, func() {
		out, err = handler.Handle(context.Background(), &GetAllDeviceDefinitionBySearchQuery{Query: "up", Page: 1, PageSize: 10})
	})
	require.NoError(t, err)
	result := out.(*GetAllDeviceDefinitionBySearchQueryResult)

	require.Len(t, result.DeviceDefinitions, 1)
	item := result.DeviceDefinitions[0]
	assert.Equal(t, "volkswagen_up_2019", item.ID)
	assert.Equal(t, "volkswagen_up_2019", item.DeviceDefinitionID)
	assert.Equal(t, 0, item.Year)
	assert.Equal(t, "", item.ImageURL)
	assert.Equal(t, []GetAllDeviceDefinitionFacetItem{}, result.Facets.Makes)
}
