package main

import (
	"context"
	"fmt"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func idSet(ids ...string) map[string]struct{} {
	s := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

func TestPruneOrphans_DeletesIndexedIDsMissingFromCatalog(t *testing.T) {
	ctrl := gomock.NewController(t)
	indexer := NewMockSearchIndexer(ctrl)

	indexer.EXPECT().ExportIDs(gomock.Any(), "defs").
		Return([]string{"keep_1", "gone_1", "keep_2"}, nil)
	indexer.EXPECT().DeleteDocuments(gomock.Any(), "defs", []string{"gone_1"}).Return(1, nil)

	n, err := pruneOrphans(context.Background(), indexer, "defs", idSet("keep_1", "keep_2"), false)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestPruneOrphans_SkipsWhenCatalogIsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	indexer := NewMockSearchIndexer(ctrl)
	// An empty catalog means the catalog read failed or returned nothing.
	// Pruning against it would empty the whole index, so don't even export.

	n, err := pruneOrphans(context.Background(), indexer, "defs", idSet(), false)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestPruneOrphans_RefusesBulkPruneUnlessAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	indexer := NewMockSearchIndexer(ctrl)

	indexed := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		indexed = append(indexed, fmt.Sprintf("id_%d", i))
	}
	indexer.EXPECT().ExportIDs(gomock.Any(), "defs").Return(indexed, nil)
	// DeleteDocuments must not be called.

	_, err := pruneOrphans(context.Background(), indexer, "defs", idSet("id_0"), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "allow-bulk-prune")
}

func TestPruneOrphans_AllowsBulkPruneWhenFlagged(t *testing.T) {
	ctrl := gomock.NewController(t)
	indexer := NewMockSearchIndexer(ctrl)

	indexed := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		indexed = append(indexed, fmt.Sprintf("id_%d", i))
	}
	indexer.EXPECT().ExportIDs(gomock.Any(), "defs").Return(indexed, nil)
	indexer.EXPECT().
		DeleteDocuments(gomock.Any(), "defs", gomock.Len(999)).
		Return(999, nil)

	n, err := pruneOrphans(context.Background(), indexer, "defs", idSet("id_0"), true)
	require.NoError(t, err)
	assert.Equal(t, 999, n)
}

// A definition below the index year cutoff is not an orphan: it exists in the
// catalog, it is simply not indexed going forward. Pruning it deletes a
// document that is live and searchable today.
func TestRunSearchSync_DoesNotPruneDefinitionsBelowTheIndexYearCutoff(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)
	onChain.EXPECT().PinCatalogSnapshot(gomock.Any()).Return(nil)
	onChain.EXPECT().CatalogIDs(gomock.Any()).Return([]string{"acura_mdx_2005", "acura_mdx_2020"}, nil).AnyTimes()

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Acura"},
	}, nil)
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{
			defToyota(2005, "MDX", "acura_mdx_2005"), // in the catalog, below the cutoff
			defToyota(2020, "MDX", "acura_mdx_2020"),
		}, nil)

	// Only the eligible definition is indexed.
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Len(1)).Return(nil)

	indexer.EXPECT().ExportIDs(gomock.Any(), "dd-search").
		Return([]string{"acura_mdx_2005", "acura_mdx_2020", "acura_gone_2019"}, nil)
	// Only the id absent from the catalog is deleted.
	indexer.EXPECT().DeleteDocuments(gomock.Any(), "dd-search", []string{"acura_gone_2019"}).
		Return(1, nil)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.NoError(t, err)
}

// identity-api returns HTTP 200 with a populated `errors` array and null data
// when it fails, and the client surfaces that as (nil, nil). Treating zero
// manufacturers as success means the daily cron reports "Index Updated" having
// indexed nothing -- and the prune pass then has an empty keep-set.
func TestRunSearchSync_FailsWhenIdentityReturnsNoManufacturers(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)
	onChain.EXPECT().PinCatalogSnapshot(gomock.Any()).Return(nil)

	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{}, nil)
	// No upserts, no export, no deletes.

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no manufacturers")
}

// The sync pins one manifest before reading the catalog. Not for freshness --
// there is no edge cache in front of the worker -- but so the per-manufacturer
// paging below cannot span two manifests mid-run.
func TestRunSearchSync_PinsAFreshCatalogSnapshotBeforePruning(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	pinned := onChain.EXPECT().PinCatalogSnapshot(gomock.Any()).Return(nil)
	identity.EXPECT().GetManufacturers().Return([]coremodels.Manufacturer{
		{TokenID: 1, Name: "Honda"},
	}, nil).After(pinned)
	// The keep-set is read from the pinned snapshot, not from the walk.
	onChain.EXPECT().CatalogIDs(gomock.Any()).Return([]string{"h1"}, nil).After(pinned)
	onChain.EXPECT().QueryDefinitionsByManufacturer(gomock.Any(), 1, 0).
		Return([]coremodels.DeviceDefinitionTablelandModel{defToyota(2020, "Civic", "h1")}, nil)
	indexer.EXPECT().UpsertDocuments(gomock.Any(), "dd-search", gomock.Any()).Return(nil)
	indexer.EXPECT().ExportIDs(gomock.Any(), "dd-search").Return([]string{"h1"}, nil)

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.NoError(t, err)
}

func TestRunSearchSync_FailsWhenTheSnapshotCannotBePinned(t *testing.T) {
	ctrl := gomock.NewController(t)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	onChain := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	indexer := NewMockSearchIndexer(ctrl)

	onChain.EXPECT().PinCatalogSnapshot(gomock.Any()).Return(fmt.Errorf("catalog returned 502"))
	// Nothing else may run: a keep-set built on a stale manifest deletes.

	err := runSearchSync(context.Background(), identity, onChain, indexer, "dd-search", false)
	require.Error(t, err)
}

// Pins the cap. One number, so there is no regime where a different constant
// silently takes over -- which is how a change to the old fraction shipped in a
// commit message and not in the code.
func TestPruneCapSitsBetweenChurnAndCatastrophe(t *testing.T) {
	// Prod carries 41 real orphans; routine churn must never trip the guard.
	assert.Greater(t, maxPruneWithoutFlag, 41)
	// A keep-set that lost its largest manufacturer is ~1,615 documents.
	assert.Less(t, maxPruneWithoutFlag, 1615)
}
