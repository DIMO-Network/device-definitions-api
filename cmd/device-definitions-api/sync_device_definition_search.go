package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"time"
)

// syncDeviceDefinitionSearchCmd cli command to sync to typesense
type syncDeviceDefinitionSearchCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	createIndex bool
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

	client := typesense.NewClient(
		typesense.WithServer(p.settings.SearchServiceAPIURL.String()),
		typesense.WithAPIKey(p.settings.SearchServiceAPIKey))

	collectionName := p.settings.SearchServiceIndexName

	if p.createIndex {

		_, err := client.Collection(collectionName).Delete(context.Background())
		if err != nil {
			p.logger.Error().Err(err).Send()
		}

		hasFacet := true
		schema := &api.CollectionSchema{
			Name: collectionName,
			Fields: []api.Field{
				{
					Name: "id",
					Type: "string",
				},
				{
					Name: "device_definition_id",
					Type: "string",
				},
				{
					Name: "name",
					Type: "string",
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
					Name:  "manufacturer_token_id",
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
			},
			DefaultSortingField: pointer.String("score"),
		}
		_, err = client.Collections().Create(context.Background(), schema)
		if err != nil {
			p.logger.Error().Err(err).Send()
			return subcommands.ExitFailure
		}

		fmt.Printf("Index %s created", collectionName)
	}

	fmt.Printf("Starting processing definitions\n")

	all, err := models.DeviceDefinitions(qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.Images),
		models.DeviceDefinitionWhere.Verified.EQ(true)).All(ctx, pdb.DBS().Reader)
	if err != nil {
		p.logger.Fatal().Err(err).Send()
	}

	var documents []SearchEntryItem
	// todo this should come from tableland - problem is iterating over all the tables, or do it via identity-api
	for _, dd := range all {
		id := dd.NameSlug
		deviceDefinitionID := dd.ID
		name := common.BuildDeviceDefinitionName(dd.Year, dd.R.DeviceMake.Name, dd.Model)
		imageUrl := ""
		for _, image := range dd.R.Images {
			imageUrl = image.SourceURL
			break
		}
		modelName := dd.Model
		modelSlug := dd.ModelSlug

		year := dd.Year
		if year < 2005 {
			continue // do not sync old cars into search
		}

		makeName := dd.R.DeviceMake.Name
		makeSlug := dd.R.DeviceMake.NameSlug
		manufacturerTokenID := int64(0)
		if !dd.R.DeviceMake.TokenID.IsZero() {
			manufacturerTokenID, _ = dd.R.DeviceMake.TokenID.Int64()
		}

		newDocument := SearchEntryItem{
			ID:                  id,
			DeviceDefinitionID:  deviceDefinitionID,
			DefinitionID:        dd.NameSlug,
			Name:                name,
			Make:                makeName,
			MakeSlug:            makeSlug,
			ManufacturerTokenID: int(manufacturerTokenID),
			Model:               modelName,
			ModelSlug:           modelSlug,
			Year:                int(year),
			ImageURL:            imageUrl,
			Score:               1,
		}

		documents = append(documents, newDocument)
	}

	err = uploadWithApi(client, documents, p.settings.SearchServiceIndexName)

	fmt.Print("Index Updated")
	return subcommands.ExitSuccess
}

func uploadWithApi(client *typesense.Client, entries []SearchEntryItem, collectionName string) error {
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
