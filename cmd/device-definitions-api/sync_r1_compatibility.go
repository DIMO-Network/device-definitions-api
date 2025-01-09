//nolint:tagliatelle
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
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

// nolint
func (p *syncR1CompatibiltyCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	compatSheetAPI := "https://device-definitions-api.dimo.zone/compatibility/r1-sheet"

	client := typesense.NewClient(
		typesense.WithServer("https://i0bj2htyg7r4l31kp.a1.typesense.net"),
		typesense.WithAPIKey("MqHDdccLzBy4tnw4fxXl12huGQMYzpjB"))

	// get optional Make filter from cmd line
	if len(p.oemFilter) > 1 {
		fmt.Println("Make filter used: " + p.oemFilter)
	}
	// Step 1: Fetch Sheety Data
	sheetyData, err := fetchSheetData(compatSheetAPI)
	if err != nil {
		p.logger.Fatal().Msgf("Error fetching Sheety API data: %v", err)
	}
	fmt.Printf("Fetched %d records\n", len(sheetyData))

	searchEntries := make([]R1SearchEntryItem, 0)

	// Step 2: Check each definitionId via GraphQL
	processedCount := 0
	for _, item := range sheetyData {
		if p.oemFilter != "" {
			if p.oemFilter != item.Make {
				continue
			}
		}
		processedCount++

		entry := R1SearchEntryItem{
			DefinitionID: item.DefinitionID,
			Make:         item.Make,
			Model:        item.Model,
			Year:         item.Year,
			Compatible:   item.Compatible,
			Name:         fmt.Sprintf("%s %s %d", item.Make, item.Model, item.Year),
		}

		searchEntries = append(searchEntries, entry)

		if processedCount%100 == 0 {
			fmt.Printf("Processed %d definitionIds\n", processedCount)
		}
	}
	fmt.Printf("Processed %d definitionIds. Uploading items...\n", processedCount)
	err = uploadR1EntriesWithAPI(client, searchEntries)
	if err != nil {
		p.logger.Fatal().Msgf("Error uploading to Typesense: %v", err)
	}

	p.logger.Info().Msg("completed syncing ruptela compatibility search")
	return subcommands.ExitSuccess
}

type R1Definition struct {
	DefinitionID string `json:"definitionId"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Compatible   string `json:"compatible"`
}

// UnmarshalJSON Custom unmarshaller for Vehicle struct
func (v *R1Definition) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with Model as interface{} to handle both types
	type Alias R1Definition
	temp := &struct {
		Model interface{} `json:"model"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}

	// Unmarshal into the temporary struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Handle the model field depending on its type
	switch model := temp.Model.(type) {
	case string:
		v.Model = model
	case float64:
		v.Model = strconv.Itoa(int(model)) // Convert number to string
	default:
		v.Model = "" // Or handle any unexpected type here
	}

	return nil
}

type R1SearchEntryItem struct {
	DefinitionID string `json:"definition_id"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Compatible   string `json:"compatible"`
	Name         string `json:"name"`
}

// fetchSheetData gets the data from the api endpoint that pulls from the google sheet
func fetchSheetData(url string) ([]R1Definition, error) {
	var result []R1Definition

	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	return result, err
}

func uploadR1EntriesWithAPI(client *typesense.Client, entries []R1SearchEntryItem) error {
	processedCount := 0
	for _, entry := range entries {
		processedCount++
		_, err := client.Collection(typeSenseR1Index).Documents().Upsert(context.Background(), entry)
		if err != nil {
			fmt.Printf("Error uploading entry: %v\n Retrying...", err)
			time.Sleep(1000)
			_, err = client.Collection(typeSenseR1Index).Documents().Upsert(context.Background(), entry)
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
