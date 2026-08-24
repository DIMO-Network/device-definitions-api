package main

import (
	"context"
	"fmt"
	"testing"

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
	indexer.EXPECT().DeleteDocuments(gomock.Any(), "defs", []string{"gone_1"}).Return(nil)

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
		Return(nil)

	n, err := pruneOrphans(context.Background(), indexer, "defs", idSet("id_0"), true)
	require.NoError(t, err)
	assert.Equal(t, 999, n)
}
