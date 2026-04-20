package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

//go:generate mockgen -source search_indexer.go -destination search_indexer_mock_test.go -package main
type SearchIndexer interface {
	RecreateIndex(ctx context.Context, schema *api.CollectionSchema) error
	UpsertDocuments(ctx context.Context, collectionName string, docs []SearchEntryItem) error
}

type typesenseSearchIndexer struct {
	client *typesense.Client
}

func NewTypesenseSearchIndexer(client *typesense.Client) SearchIndexer {
	return &typesenseSearchIndexer{client: client}
}

func (t *typesenseSearchIndexer) RecreateIndex(ctx context.Context, schema *api.CollectionSchema) error {
	if _, err := t.client.Collection(schema.Name).Delete(ctx); err != nil {
		// Deletion failure is non-fatal (index may not exist yet) — caller logs.
		fmt.Printf("RecreateIndex: delete returned %v (continuing)\n", err)
	}
	if _, err := t.client.Collections().Create(ctx, schema); err != nil {
		return errors.Wrap(err, "failed to create collection")
	}
	return nil
}

func (t *typesenseSearchIndexer) UpsertDocuments(ctx context.Context, collectionName string, docs []SearchEntryItem) error {
	if len(docs) == 0 {
		return nil
	}
	payload := make([]interface{}, 0, len(docs))
	for _, d := range docs {
		payload = append(payload, d)
	}
	action := "upsert"
	_, err := t.client.Collection(collectionName).Documents().Import(ctx, payload, &api.ImportDocumentsParams{
		Action: &action,
	})
	if err != nil {
		return errors.Wrap(err, "failed to import documents")
	}
	return nil
}
