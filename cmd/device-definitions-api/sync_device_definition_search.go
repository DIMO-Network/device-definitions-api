package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	models2 "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
)

// syncDeviceDefinitionSearchCmd cli command to sync to typesense
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

// nolint
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
	//queryInstance, err := contracts.NewRegistry(p.settings.EthereumRegistryAddress, ethClient)
	//if err != nil {
	//	p.logger.Fatal().Err(err).Msg("Failed to create registry query instance.")
	//}

	onChainSvc := gateways.NewDeviceDefinitionOnChainService(&p.settings, &p.logger, ethClient, chainID, p.sender, pdb.DBS)

	client := typesense.NewClient(
		typesense.WithServer(p.settings.SearchServiceAPIURL.String()),
		typesense.WithAPIKey(p.settings.SearchServiceAPIKey))

	collectionName := p.settings.SearchServiceIndexName

	if p.createIndex {

		_, err := client.Collection(collectionName).Delete(context.Background())
		if err != nil {
			p.logger.Error().Err(err).Send()
		}
		fmt.Println("Successfully deleted index: " + collectionName)

		hasFacet := true
		nestedFields := false
		sort := true
		schema := &api.CollectionSchema{
			Name:                collectionName,
			EnableNestedFields:  &nestedFields,
			DefaultSortingField: pointer.String("score"),
			Fields: []api.Field{
				{
					Name: "device_definition_id",
					Type: "string",
				},
				{
					Name: "name",
					Type: "string",
					Sort: &sort,
				},
				{
					Name:  "make",
					Type:  "string",
					Facet: &hasFacet,
				},
				{
					Name:  "make_slug",
					Type:  "string",
					Facet: &hasFacet,
				},
				{
					Name:  "make_token_id",
					Type:  "int32",
					Facet: &hasFacet,
				},
				{
					Name:  "model",
					Type:  "string",
					Facet: &hasFacet,
				},
				{
					Name:  "model_slug",
					Type:  "string",
					Facet: &hasFacet,
				},
				{
					Name:  "year",
					Type:  "int32",
					Facet: &hasFacet,
				},
				{
					Name: "image_url",
					Type: "string",
				},
				{
					Name: "score",
					Type: "int32",
				},
				{
					Name: "definition_id",
					Type: "string",
				},
			},
		}
		_, err = client.Collections().Create(context.Background(), schema)
		if err != nil {
			p.logger.Error().Err(err).Send()
			return subcommands.ExitFailure
		}

		fmt.Printf("Index %s created\n", collectionName)
	}

	fmt.Printf("Starting processing definitions\n")

	makes, err := models.DeviceMakes().All(ctx, pdb.DBS().Reader)
	if err != nil {
		p.logger.Fatal().Err(err).Send()
	}

	var documents []SearchEntryItem
	// iterate over all makes, then query tableland
	for _, dm := range makes {
		manufacturer, err := onChainSvc.GetManufacturer(dm.NameSlug)
		if err != nil {
			p.logger.Fatal().Err(err).Send()
		}

		definitions, err := fetchAllDefinitions(ctx, onChainSvc, manufacturer.TokenID, "")
		if err != nil {
			p.logger.Fatal().Err(err).Send()
		}

		for _, dd := range definitions {
			id := dd.ID
			deviceDefinitionID := dd.ID
			name := common.BuildDeviceDefinitionName(int16(dd.Year), dm.Name, dd.Model)
			imageUrl := dd.ImageURI
			modelName := dd.Model
			modelSlug := stringutils.SlugString(dd.Model)

			year := dd.Year
			if year < 2007 {
				continue // do not sync old cars into search
			}

			makeName := dm.Name
			makeSlug := dm.NameSlug
			manufacturerTokenID := manufacturer.TokenID

			newDocument := SearchEntryItem{
				ID:                  id,
				DeviceDefinitionID:  deviceDefinitionID,
				DefinitionID:        dd.ID,
				Name:                name,
				Make:                makeName,
				MakeSlug:            makeSlug,
				ManufacturerTokenID: manufacturerTokenID,
				Model:               modelName,
				ModelSlug:           modelSlug,
				Year:                year,
				ImageURL:            imageUrl,
				Score:               1,
			}

			documents = append(documents, newDocument)
		}
	}

	err = uploadWithAPI(client, documents, p.settings.SearchServiceIndexName)

	fmt.Print("Index Updated")
	return subcommands.ExitSuccess
}

func fetchAllDefinitions(ctx context.Context, onChainSvc gateways.DeviceDefinitionOnChainService, manufacturerID int, whereClause string) ([]models2.DeviceDefinitionTablelandModel, error) {
	pageIndex := 0
	var allDefinitions []models2.DeviceDefinitionTablelandModel

	for {
		definitions, err := onChainSvc.QueryDefinitionsCustom(ctx, manufacturerID, whereClause, pageIndex)
		if err != nil {
			return nil, err
		}

		// Append the current page of definitions to allDefinitions.
		allDefinitions = append(allDefinitions, definitions...)

		// If you receive less than 50 results then you've likely reached the end.
		if len(definitions) < 50 {
			break
		}

		// Move to the next page.
		pageIndex++
	}

	return allDefinitions, nil
}

func uploadWithAPI(client *typesense.Client, entries []SearchEntryItem, collectionName string) error {
	processedCount := 0
	for _, entry := range entries {
		processedCount++
		_, err := client.Collection(collectionName).Documents().Upsert(context.Background(), entry)
		if err != nil {
			fmt.Printf("Error uploading entry: %v\n Retrying...", err)
			time.Sleep(1000)
			_, err = client.Collection(collectionName).Documents().Upsert(context.Background(), entry)
			// todo fancier retry
			if err != nil {
				return err
			}
		}
		if processedCount%100 == 0 {
			fmt.Printf("Uploaded %d definitionIds to Typesense search.\n", processedCount)
		}
	}
	return nil
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
