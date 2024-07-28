package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
)

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
		typesense.WithServer(p.settings.SearchServiceAPIURL),
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

	url := "https://device-definitions-api.dimo.zone/device-definitions/all"

	resp, err := http.Get(url)
	if err != nil {
		p.logger.Error().Err(err).Send()
		return subcommands.ExitFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error().Err(err).Send()
		return subcommands.ExitFailure
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error().Err(err).Send()
		return subcommands.ExitFailure
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		p.logger.Error().Err(err).Send()
		return subcommands.ExitFailure
	}

	var documents []interface{}

	if deviceDefinitions, ok := result["device_definitions"].([]interface{}); ok {
		for _, dd := range deviceDefinitions {
			if deviceMap, ok := dd.(map[string]interface{}); ok {
				id := deviceMap["name_slug"].(string)
				deviceDefinitionID := deviceMap["device_definition_id"].(string)
				name := deviceMap["name"].(string)

				typeMap := deviceMap["type"].(map[string]interface{})
				modelName := typeMap["model"].(string)
				modelSlug := typeMap["model_slug"].(string)

				var year int
				if yearFloat, ok := typeMap["year"].(float64); ok {
					year = int(yearFloat)
				} else {
					continue
				}

				makeMap := deviceMap["make"].(map[string]interface{})
				makeName := makeMap["name"].(string)
				makeSlug := makeMap["name_slug"].(string)

				newDocument := struct {
					ID                 string `json:"id"`
					DeviceDefinitionID string `json:"device_definition_id"` //nolint
					Name               string `json:"name"`
					Make               string `json:"make"`
					MakeSlug           string `json:"make_slug"` //nolint
					Model              string `json:"model"`
					ModelSlug          string `json:"model_slug"` //nolint
					Year               int    `json:"year"`
					ImageURL           string `json:"image_url"` //nolint
					Score              int    `json:"score"`
				}{
					ID:                 id,
					DeviceDefinitionID: deviceDefinitionID,
					Name:               name,
					Make:               makeName,
					MakeSlug:           makeSlug,
					Model:              modelName,
					ModelSlug:          modelSlug,
					Year:               year,
					ImageURL:           "",
					Score:              1,
				}

				documents = append(documents, newDocument)
			}
		}
	}

	batchSize := 100
	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		_, err = client.Collection(collectionName).
			Documents().
			Import(context.Background(), batch, &api.ImportDocumentsParams{})

		if err != nil {
			p.logger.Error().Err(err).Send()
			return subcommands.ExitFailure
		} else {
			fmt.Printf("Documents imported successfully: %d - %d\n", i, end)
		}
	}

	fmt.Print("Index Updated")
	return subcommands.ExitSuccess
}
