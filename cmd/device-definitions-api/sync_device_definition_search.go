package main

import (
	"context"
	"flag"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/ethereum/go-ethereum/ethclient"

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
	minSearchYear         = 2007
	tablelandPageSize     = 500
	searchDefaultScore    = 1
)

type syncDeviceDefinitionSearchCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	createIndex bool
	sender      sender.Sender
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
}

func (p *syncDeviceDefinitionSearchCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	ethClient, err := ethclient.Dial(p.settings.EthereumRPCURL.String())
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}

	onChainSvc := gateways.NewDeviceDefinitionOnChainService(&p.settings, &p.logger, ethClient, chainID, p.sender, pdb.DBS)
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

	if err := runSearchSync(ctx, identity, onChainSvc, indexer, collectionName); err != nil {
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
	onChainSvc gateways.DeviceDefinitionOnChainService,
	indexer SearchIndexer,
	collectionName string,
) error {
	makes, err := identity.GetManufacturers()
	if err != nil {
		return fmt.Errorf("get manufacturers: %w", err)
	}
	fmt.Printf("Found %d manufacturers\n", len(makes))

	for _, dm := range makes {
		docs, err := buildManufacturerDocuments(ctx, onChainSvc, dm)
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
	return nil
}

// buildManufacturerDocuments pulls every tableland definition for a manufacturer
// and converts the ones from model year >= minSearchYear into SearchEntryItems.
func buildManufacturerDocuments(
	ctx context.Context,
	onChainSvc gateways.DeviceDefinitionOnChainService,
	dm coremodels.Manufacturer,
) ([]SearchEntryItem, error) {
	makeSlug := stringutils.SlugString(dm.Name)
	var docs []SearchEntryItem

	pageIndex := 0
	for {
		page, err := onChainSvc.QueryDefinitionsCustom(ctx, dm.TokenID, "", pageIndex)
		if err != nil {
			return nil, err
		}
		for _, dd := range page {
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
		if len(page) < tablelandPageSize {
			break
		}
		pageIndex++
	}
	return docs, nil
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
