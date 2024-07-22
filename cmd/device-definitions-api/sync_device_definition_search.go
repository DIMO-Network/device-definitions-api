package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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

	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		models.DeviceDefinitionWhere.Year.GTE(2012),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, pdb.DBS().Reader)

	if err != nil {
		p.logger.Error().Err(err).Send()
		return subcommands.ExitFailure
	}
	fmt.Printf("Found %d device-definition(s) in all device-definitions", len(all))

	for _, dd := range all {
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
			ID:                 dd.NameSlug,
			DeviceDefinitionID: dd.ID,
			Name:               common.BuildDeviceDefinitionName(dd.Year, dd.R.DeviceMake.Name, dd.Model),
			Make:               dd.R.DeviceMake.Name,
			MakeSlug:           dd.R.DeviceMake.NameSlug,
			Model:              dd.Model,
			ModelSlug:          dd.ModelSlug,
			Year:               int(dd.Year),
			ImageURL:           ResolveImageURL(dd),
			Score:              1,
		}

		_, err = client.Collection(collectionName).Documents().Upsert(context.Background(), newDocument)
		if err != nil {
			p.logger.Error().Err(err).Send()
		}

		fmt.Printf("Document Updated => %s \n", newDocument.Name)
	}

	fmt.Print("Index Updated")
	return subcommands.ExitSuccess
}

func ResolveImageURL(dd *models.DeviceDefinition) string {
	img := ""
	if dd.R.Images != nil {
		w := 0
		for _, image := range dd.R.Images {
			extra := 0
			if !image.NotExactImage {
				extra = 2000 // we want to give preference to exact images
			}
			if image.Width.Int+extra > w {
				w = image.Width.Int + extra
				img = image.SourceURL
			}
		}
	}
	return img
}
