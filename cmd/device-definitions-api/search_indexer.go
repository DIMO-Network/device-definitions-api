package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
)

//go:generate mockgen -source search_indexer.go -destination search_indexer_mock_test.go -package main
type SearchIndexer interface {
	RecreateIndex(ctx context.Context, schema *api.CollectionSchema) error
	UpsertDocuments(ctx context.Context, collectionName string, docs []SearchEntryItem) error
	// ExportIDs returns the id of every document currently in the collection.
	ExportIDs(ctx context.Context, collectionName string) ([]string, error)
	// DeleteDocuments removes the given ids and reports how many it actually
	// removed, so a partial failure does not report ids that still exist.
	DeleteDocuments(ctx context.Context, collectionName string, ids []string) (int, error)
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

func (t *typesenseSearchIndexer) ExportIDs(ctx context.Context, collectionName string) ([]string, error) {
	body, err := t.client.Collection(collectionName).Documents().Export(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to export documents")
	}
	defer body.Close() //nolint:errcheck

	var ids []string
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var doc struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(line, &doc); err != nil {
			return nil, errors.Wrap(err, "failed to decode exported document")
		}
		if doc.ID != "" {
			ids = append(ids, doc.ID)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read the export stream")
	}
	return ids, nil
}

// DeleteDocuments removes ids one at a time rather than through a filter_by
// set: definition ids legitimately contain & + ( ) and ", which would need
// escaping inside a filter expression. Orphans are rare, so this stays cheap.
func (t *typesenseSearchIndexer) DeleteDocuments(ctx context.Context, collectionName string, ids []string) (int, error) {
	deleted := 0
	for _, id := range ids {
		if _, err := t.client.Collection(collectionName).Document(id).Delete(ctx); err != nil {
			return deleted, errors.Wrapf(err, "failed to delete document %s", id)
		}
		deleted++
		fmt.Printf("  pruned search document with no definition: %s\n", id)
	}
	return deleted, nil
}
