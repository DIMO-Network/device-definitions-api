//nolint:tagliatelle
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/google/subcommands"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
)

const typeSenseR1Index = "r1_compatibility"

// syncR1CompatibiltyCmd cli command to sync from google spreadsheet to Typesense search index for R1 compatibilty
type syncR1CompatibiltyCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	createIndex bool
	oemFilter   string
}

func (*syncR1CompatibiltyCmd) Name() string { return "sync-r1-compatibilty" }
func (*syncR1CompatibiltyCmd) Synopsis() string {
	return "sync r1 google spreadsheet to typesense search"
}
func (*syncR1CompatibiltyCmd) Usage() string {
	return `sync-r1-compatibilty`
}

func (p *syncR1CompatibiltyCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.createIndex, "create-index", false, "create or recreate index")
	f.StringVar(&p.oemFilter, "oem-filter", "", "oem filter")
}

func (p *syncR1CompatibiltyCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// get data from sheet using google sheets
	qh := queries.NewCompatibilityR1SheetQueryHandler(&p.settings)
	compatibilityRows, err := qh.Handle(ctx, nil)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("error fetching compatibility sheet data")
		return subcommands.ExitFailure
	}
	sheetData := compatibilityRows.([]queries.CompatibilitySheetRow)

	fmt.Printf("Fetched %d records\n", len(sheetData))

	client := typesense.NewClient(
		typesense.WithServer(p.settings.SearchServiceAPIURL.String()),
		typesense.WithAPIKey(p.settings.SearchServiceAPIKey))

	// get optional Make filter from cmd line
	if len(p.oemFilter) > 1 {
		fmt.Println("Make filter used: " + p.oemFilter)
	}
	searchEntries := make([]queries.GetR1SearchEntryItem, 0)

	// Step 2: Check each definitionId via GraphQL
	processedCount := 0
	for _, item := range sheetData {
		if p.oemFilter != "" {
			if p.oemFilter != item.Make {
				continue
			}
		}
		processedCount++

		entry := queries.GetR1SearchEntryItem{
			DefinitionID: item.DefinitionID,
			Make:         item.Make,
			Model:        item.Model,
			Year:         item.Year,
			Compatible:   item.Compatible,
			Name:         fmt.Sprintf("%s %s %d", item.Make, item.Model, item.Year),
		}

		searchEntries = append(searchEntries, entry)
	}
	fmt.Printf("Processed %d definitionIds. Uploading items...\n", processedCount)
	if p.createIndex {
		err := createR1CompatibilityIndex(p.logger, client, typeSenseR1Index)
		if err != nil {
			p.logger.Fatal().Msgf("error creating index: %v", err)
		}
		p.logger.Info().Msg("index created: " + typeSenseR1Index)
	}
	err = uploadR1EntriesWithAPI(ctx, client, searchEntries)
	if err != nil {
		p.logger.Fatal().Msgf("error uploading to Typesense: %v", err)
	}

	p.logger.Info().Msg("completed syncing ruptela compatibility search")
	return subcommands.ExitSuccess
}

func uploadR1EntriesWithAPI(ctx context.Context, client *typesense.Client, entries []queries.GetR1SearchEntryItem) error {
	//processedCount := 0
	action := "upsert"
	var interfaceSlice []interface{}
	for _, entry := range entries {
		// some validation
		if entry.DefinitionID != "" && entry.Make != "" && entry.Model != "" && entry.Name != "" && entry.Compatible != "" {
			interfaceSlice = append(interfaceSlice, entry)
		}
	}

	responses, err := client.Collection(typeSenseR1Index).Documents().Import(ctx, interfaceSlice, &api.ImportDocumentsParams{
		Action:                   &action,
		BatchSize:                nil,
		DirtyValues:              nil,
		RemoteEmbeddingBatchSize: nil,
	})
	if err != nil {
		return errors.Wrap(err, "failed to import documents")
	}

	fmt.Printf("Uploaded %d definitions to Typesense search.\n", len(responses))

	//for _, entry := range entries {
	//	processedCount++
	//
	//	_, err := client.Collection(typeSenseR1Index).Documents().Upsert(ctx, entry)
	//	if err != nil {
	//		fmt.Printf("Error uploading entry: %v\n Retrying...", err)
	//		time.Sleep(1000)
	//		_, err = client.Collection(typeSenseR1Index).Documents().Upsert(ctx, entry)
	//		// todo fancier retry
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	if processedCount%100 == 0 {
	//		fmt.Printf("Uploaded %d definitionIds to Typesense search.\n", processedCount)
	//	}
	//}
	return nil
}

func createR1CompatibilityIndex(logger zerolog.Logger, client *typesense.Client, collectionName string) error {
	_, err := client.Collection(collectionName).Delete(context.Background())
	if err != nil {
		logger.Error().Err(err).Send()
	}
	fmt.Println("Successfully deleted index: " + collectionName)

	hasFacet := true
	nestedFields := false
	sort := true
	schema := &api.CollectionSchema{
		Name:                collectionName,
		EnableNestedFields:  &nestedFields,
		DefaultSortingField: pointer.String("year"),
		Fields: []api.Field{
			{
				// this will hold the device_definition_id - must be called id for typesense upsert to work
				Name: "id",
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
				Name:  "model",
				Type:  "string",
				Facet: &hasFacet,
			},
			{
				Name:  "year",
				Type:  "int32",
				Facet: &hasFacet,
			},
			{
				Name: "compatible",
				Type: "string",
			},
		},
	}
	_, err = client.Collections().Create(context.Background(), schema)
	if err != nil {
		return errors.Wrap(err, "failed to create collection")
	}
	return nil
}
