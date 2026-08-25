package main

import (
	"context"
	"flag"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
)

const (
	minSearchYear = 2007
	// Most search documents the prune will delete in one run without
	// -allow-bulk-prune. Prod carries 41 orphans; a keep-set that lost a whole
	// manufacturer is ~1,615 documents. Anything in between means the keep-set
	// is wrong, not that the catalog shrank.
	maxPruneWithoutFlag = 500
	searchDefaultScore  = 1
)

type syncDeviceDefinitionSearchCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	createIndex    bool
	allowBulkPrune bool
}

func (*syncDeviceDefinitionSearchCmd) Name() string { return "sync-device-definitions-search" }
func (*syncDeviceDefinitionSearchCmd) Synopsis() string {
	return "sync device definitions search index"
}
func (*syncDeviceDefinitionSearchCmd) Usage() string {
	return `sync-device-definitions-search`
}

func (p *syncDeviceDefinitionSearchCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.createIndex, "create-index", false, "create or recreate index")
	f.BoolVar(&p.allowBulkPrune, "allow-bulk-prune", false,
		"permit deleting an unusually large share of the search index (normally a sign of a partial catalog read)")
}

func (p *syncDeviceDefinitionSearchCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	catalogSvc := gateways.NewDeviceDefinitionCatalogService(&p.settings, &p.logger)
	identity := gateways.NewIdentityAPIService(&p.logger, &p.settings)

	client := typesense.NewClient(
		typesense.WithServer(p.settings.SearchServiceAPIURL.String()),
		typesense.WithAPIKey(p.settings.SearchServiceAPIKey))
	indexer := NewTypesenseSearchIndexer(client)

	collectionName := p.settings.SearchServiceIndexName

	if p.createIndex {
		if err := indexer.RecreateIndex(ctx, deviceDefinitionSearchSchema(collectionName)); err != nil {
			p.logger.Error().Err(err).Send()
			return subcommands.ExitFailure
		}
		fmt.Printf("Index %s created\n", collectionName)
	}

	if err := runSearchSync(ctx, identity, catalogSvc, indexer, collectionName, p.allowBulkPrune); err != nil {
		p.logger.Error().Err(err).Msg("sync failed")
		return subcommands.ExitFailure
	}
	fmt.Print("Index Updated")
	return subcommands.ExitSuccess
}

// runSearchSync fetches every manufacturer, builds its definition documents, and
// upserts them to the search index one manufacturer at a time. This caps the
// steady-state memory footprint to a single manufacturer's definitions rather
// than the full catalog.
func runSearchSync(
	ctx context.Context,
	identity gateways.IdentityAPI,
	catalogSvc gateways.DeviceDefinitionCatalogService,
	indexer SearchIndexer,
	collectionName string,
	allowBulkPrune bool,
) error {
	// Pin one manifest for the whole run. This is not a freshness measure: the
	// worker generates every response, so there is no edge cache, and the
	// keep-set is a single CatalogIDs read that cannot span manifests either
	// way. It keeps the per-manufacturer paging below consistent, whose failure
	// mode is a missed upsert that the next daily run repairs.
	if err := catalogSvc.PinCatalogSnapshot(ctx); err != nil {
		return fmt.Errorf("pin catalog snapshot: %w", err)
	}

	makes, err := identity.GetManufacturers()
	if err != nil {
		return fmt.Errorf("get manufacturers: %w", err)
	}
	// identity-api answers HTTP 200 with a populated `errors` array and null
	// data when it fails, which the client surfaces as (nil, nil). Without this
	// the cron prints "Index Updated" having indexed nothing, and the prune
	// pass runs with an empty keep-set.
	if len(makes) == 0 {
		return fmt.Errorf("identity returned no manufacturers; refusing to sync an empty catalog")
	}
	fmt.Printf("Found %d manufacturers\n", len(makes))

	// The keep-set comes from the manifest, not from walking manufacturers. A
	// walk is only as complete as identity-api's manufacturer list, and one
	// make missing from it silently drops all of its definitions from the
	// keep-set -- BMW alone is 1,618 documents, which slides under the bulk
	// guard's 1,669 and gets deleted without tripping anything.
	allIDs, err := catalogSvc.CatalogIDs(ctx)
	if err != nil {
		return fmt.Errorf("read catalog ids: %w", err)
	}
	if len(allIDs) == 0 {
		return fmt.Errorf("catalog returned no definitions; refusing to sync against an empty keep-set")
	}
	catalogIDs := make(map[string]struct{}, len(allIDs))
	for _, id := range allIDs {
		catalogIDs[id] = struct{}{}
	}

	for _, dm := range makes {
		docs, _, err := buildManufacturerDocuments(ctx, catalogSvc, dm)
		if err != nil {
			return fmt.Errorf("build documents for %s: %w", dm.Name, err)
		}
		if len(docs) == 0 {
			fmt.Printf("%s: no definitions to sync\n", dm.Name)
			continue
		}
		if err := indexer.UpsertDocuments(ctx, collectionName, docs); err != nil {
			return fmt.Errorf("upsert %s: %w", dm.Name, err)
		}
		fmt.Printf("%s: upserted %d definitions\n", dm.Name, len(docs))
	}

	// Upserting alone leaves a definition deleted from the catalog searchable
	// forever, returning an id that 404s.
	pruned, err := pruneOrphans(ctx, indexer, collectionName, catalogIDs, allowBulkPrune)
	if err != nil {
		// Report what was actually removed before the failure: the deletes are
		// not transactional and the removed documents do not come back.
		return fmt.Errorf("pruned %d search documents before failing: %w", pruned, err)
	}
	if pruned > 0 {
		fmt.Printf("pruned %d search documents with no definition in the catalog\n", pruned)
	}
	return nil
}

// buildManufacturerDocuments pulls every definition for a manufacturer and
// converts the ones from model year >= minSearchYear into SearchEntryItems. It
// also returns every id it saw, which callers may ignore: the prune's keep-set
// comes from the manifest via CatalogIDs, not from this walk.
func buildManufacturerDocuments(
	ctx context.Context,
	catalogSvc gateways.DeviceDefinitionCatalogService,
	dm coremodels.Manufacturer,
) ([]SearchEntryItem, []string, error) {
	makeSlug := stringutils.SlugString(dm.Name)
	var docs []SearchEntryItem
	var allIDs []string

	pageIndex := 0
	for {
		page, err := catalogSvc.QueryDefinitionsByManufacturer(ctx, dm.TokenID, pageIndex)
		if err != nil {
			return nil, nil, err
		}
		for _, dd := range page {
			allIDs = append(allIDs, dd.ID)
			if dd.Year < minSearchYear {
				continue
			}
			docs = append(docs, SearchEntryItem{
				ID:                  dd.ID,
				DeviceDefinitionID:  dd.ID,
				DefinitionID:        dd.ID,
				Name:                common.BuildDeviceDefinitionName(int16(dd.Year), dm.Name, dd.Model),
				Make:                dm.Name,
				MakeSlug:            makeSlug,
				ManufacturerTokenID: dm.TokenID,
				Model:               dd.Model,
				ModelSlug:           stringutils.SlugString(dd.Model),
				Year:                dd.Year,
				ImageURL:            dd.ImageURI,
				Score:               searchDefaultScore,
			})
		}
		if len(page) < gateways.CatalogPageSize {
			break
		}
		pageIndex++
	}
	return docs, allIDs, nil
}

// pruneOrphans deletes search documents whose definition no longer exists in
// the catalog. The sync pass only ever upserts, so without this a definition
// deleted from R2 stays searchable forever, returning an id that 404s.
func pruneOrphans(ctx context.Context, indexer SearchIndexer, collectionName string, catalogIDs map[string]struct{}, allowBulk bool) (int, error) {
	// An empty catalog set means the catalog read failed or returned nothing.
	// Pruning against it would empty the whole index.
	if len(catalogIDs) == 0 {
		return 0, nil
	}

	indexed, err := indexer.ExportIDs(ctx, collectionName)
	if err != nil {
		return 0, fmt.Errorf("export index ids: %w", err)
	}

	var stale []string
	for _, id := range indexed {
		if _, ok := catalogIDs[id]; !ok {
			stale = append(stale, id)
		}
	}
	if len(stale) == 0 {
		return 0, nil
	}

	// A prune this large is far more likely to mean the catalog read was
	// partial than that the catalog really shrank that much. Fail loud rather
	// than quietly gut the index.
	if len(stale) > maxPruneWithoutFlag && !allowBulk {
		return 0, fmt.Errorf(
			"refusing to prune %d of %d search documents (limit %d); re-run with -allow-bulk-prune if this is expected",
			len(stale), len(indexed), maxPruneWithoutFlag)
	}

	// Documents below the index year cutoff are never re-upserted by a sync, so
	// a wrong deletion here is permanent and this log is the only record of it.
	// DeleteDocuments reports how many it actually removed, so a partial
	// failure does not claim ids that still exist.
	deleted, err := indexer.DeleteDocuments(ctx, collectionName, stale)
	if err != nil {
		return deleted, fmt.Errorf("delete orphaned search documents: %w", err)
	}
	return deleted, nil
}

func deviceDefinitionSearchSchema(collectionName string) *api.CollectionSchema {
	hasFacet := true
	nestedFields := false
	sort := true
	return &api.CollectionSchema{
		Name:                collectionName,
		EnableNestedFields:  &nestedFields,
		DefaultSortingField: pointer.String("score"),
		Fields: []api.Field{
			{Name: "device_definition_id", Type: "string"},
			{Name: "name", Type: "string", Sort: &sort},
			{Name: "make", Type: "string", Facet: &hasFacet},
			{Name: "make_slug", Type: "string", Facet: &hasFacet},
			{Name: "make_token_id", Type: "int32", Facet: &hasFacet},
			{Name: "model", Type: "string", Facet: &hasFacet},
			{Name: "model_slug", Type: "string", Facet: &hasFacet},
			{Name: "year", Type: "int32", Facet: &hasFacet},
			{Name: "image_url", Type: "string"},
			{Name: "score", Type: "int32"},
			{Name: "definition_id", Type: "string"},
		},
	}
}

type SearchEntryItem struct {
	ID                  string `json:"id"`
	DeviceDefinitionID  string `json:"device_definition_id"` //nolint
	Name                string `json:"name"`
	Make                string `json:"make"`
	MakeSlug            string `json:"make_slug"`     //nolint
	ManufacturerTokenID int    `json:"make_token_id"` //nolint
	Model               string `json:"model"`
	ModelSlug           string `json:"model_slug"` //nolint
	Year                int    `json:"year"`
	ImageURL            string `json:"image_url"` //nolint
	Score               int    `json:"score"`
	DefinitionID        string `json:"definition_id"` //nolint
}
