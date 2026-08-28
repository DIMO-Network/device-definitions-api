package gateways

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Runs only against a live catalog (local wrangler dev or deployed worker):
//
//	CATALOG_E2E_URL=http://localhost:8787 go test ./internal/infrastructure/gateways/ -run TestCatalogServiceE2E
func TestCatalogServiceE2E(t *testing.T) {
	url := os.Getenv("CATALOG_E2E_URL")
	if url == "" {
		t.Skip("CATALOG_E2E_URL not set")
	}
	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: url}, &logger)
	ctx := context.Background()

	dd, manufID, err := svc.GetTemplateByID(ctx, "dodge_town-&-country_2012")
	require.NoError(t, err)
	require.NotNil(t, dd)
	assert.Equal(t, "Town & Country", dd.Model)
	assert.Equal(t, int64(33), manufID.Int64())
	require.NotEmpty(t, dd.Trims)

	// Manufacturer scoping: right owner resolves, wrong owner returns nil.
	scoped, err := svc.GetDeviceDefinitionByID(ctx, big.NewInt(33), "dodge_town-&-country_2012")
	require.NoError(t, err)
	require.NotNil(t, scoped)
	wrong, err := svc.GetDeviceDefinitionByID(ctx, big.NewInt(13), "dodge_town-&-country_2012")
	require.NoError(t, err)
	assert.Nil(t, wrong)

	// Paged listing as the Typesense sync job consumes it.
	page0, err := svc.QueryDefinitionsByManufacturer(ctx, 13, 0)
	require.NoError(t, err)
	assert.Len(t, page0, 500)
	total := len(page0)
	for i := 1; ; i++ {
		page, err := svc.QueryDefinitionsByManufacturer(ctx, 13, i)
		require.NoError(t, err)
		total += len(page)
		if len(page) < 500 {
			break
		}
	}
	assert.Greater(t, total, 1500, "BMW should have >1500 definitions")

	// No fallback to the pre-migration definitions/<id>.json: a missing
	// template must fail loudly, not resolve as a quiet nil.
	missing, _, err := svc.GetTemplateByID(ctx, "not_areal_2020")
	require.Error(t, err)
	assert.Nil(t, missing)
}
