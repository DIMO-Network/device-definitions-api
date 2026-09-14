package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The definitions-worker indexes one document per trim. Without group_by the
// endpoint would return a definition once per trim, and the handler's
// one-item-per-definition contract depends on the query being grouped -- which
// is only observable on the wire.
func TestGetDeviceDefinitionsGroupsByDefinitionID(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"found": 0, "grouped_hits": [], "facet_counts": []}`))
	}))
	defer srv.Close()

	apiURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	logger := zerolog.Nop()
	svc := NewTypesenseAPIService(&config.Settings{
		SearchServiceAPIURL:    *apiURL,
		SearchServiceAPIKey:    "test-key",
		SearchServiceIndexName: "definitions_prod",
	}, &logger)

	result, err := svc.GetDeviceDefinitions(context.Background(), "camry", "toyota", "", 2020, 1, 20)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "/collections/definitions_prod/documents/search", gotPath)
	assert.Equal(t, "definition_id", gotQuery.Get("group_by"))
	assert.Equal(t, "1", gotQuery.Get("group_limit"))
	assert.Equal(t, "make,model,year", gotQuery.Get("facet_by"))
	assert.Equal(t, "name", gotQuery.Get("query_by"))
	assert.Equal(t, "make_slug:=toyota && year:=2020", gotQuery.Get("filter_by"))
}
