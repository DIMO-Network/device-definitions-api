package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type syncFuelImageCmd struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*syncFuelImageCmd) Name() string     { return "images" }
func (*syncFuelImageCmd) Synopsis() string { return "images args to stdout." }
func (*syncFuelImageCmd) Usage() string {
	return `images [] <some text>:
	images args.
  `
}

func (p *syncFuelImageCmd) SetFlags(_ *flag.FlagSet) {

}

func (p *syncFuelImageCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := fetchFuelAPIImages(ctx, p.logger, &p.settings)
	if err != nil {
		p.logger.Error().Err(err)
	}
	return subcommands.ExitSuccess
}

type deviceData struct {
	Make   string
	Models []model
}

type model struct {
	Model              string
	Year               int
	DeviceDefinitionID string
}

func fetchFuelAPIImages(ctx context.Context, logger zerolog.Logger, settings *config.Settings) error {
	pdb := db.NewDbConnectionFromSettings(ctx, &settings.DB, true)
	pdb.WaitForDB(logger)

	fs := gateways.NewFuelAPIService(settings, &logger)
	devices, err := getDeviceData(ctx, pdb)
	if err != nil {
		return err
	}
	logger.Info().Msgf("pulling fuel images for %d device definitions", len(devices))

	err = writeToTable(ctx, pdb, logger, fs, devices, 2, 2)
	if err != nil {
		logger.Err(err).Msg("failed to writeToTable when fetching Fuel API images")
	}
	err = writeToTable(ctx, pdb, logger, fs, devices, 2, 6)
	if err != nil {
		logger.Err(err).Msg("failed to writeToTable when fetching Fuel API images")
	}

	return nil
}

func writeToTable(ctx context.Context, store db.Store, logger zerolog.Logger, fs gateways.FuelAPIService, data []deviceData, prodID int, prodFormat int) error {

	for _, d := range data {
		for n := range d.Models {
			img, err := fs.FetchDeviceImages(d.Make, d.Models[n].Model, d.Models[n].Year, prodID, prodFormat)
			if err != nil {
				logger.Warn().Msgf("unable to fetch device image for: %d %s %s", d.Models[n].Year, d.Make, d.Models[n].Model)
				continue
			}
			var p models.Image

			// loop through all img (color variations)
			for _, device := range img.Images {
				p.ID = ksuid.New().String()
				p.DeviceDefinitionID = d.Models[n].DeviceDefinitionID
				p.FuelAPIID = null.StringFrom(img.FuelAPIID)
				p.Width = null.IntFrom(img.Width)
				p.Height = null.IntFrom(img.Height)
				p.SourceURL = device.SourceURL
				//p.DimoS3URL = null.StringFrom("") // dont set it so it is null
				p.Color = device.Color
				p.NotExactImage = img.NotExactImage

				err = p.Upsert(ctx, store.DBS().Writer, true, []string{models.ImageColumns.DeviceDefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// deviceData looks for makes and models in our database and returns a projection of them specific to Fuel
func getDeviceData(ctx context.Context, d db.Store) ([]deviceData, error) {

	oems, err := models.DeviceMakes().All(ctx, d.DBS().Reader)
	if err != nil {
		return []deviceData{}, err
	}

	devices := make([]deviceData, len(oems))
	for n, mk := range oems {
		mdls, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(mk.ID),
			models.DeviceDefinitionWhere.Year.GTE(2005)).All(ctx, d.DBS().Reader)
		if err != nil {
			return []deviceData{}, err
		}
		devices[n] = deviceData{Make: mk.NameSlug, Models: make([]model, len(mdls))}
		for i, mdl := range mdls {
			devices[n].Models[i] = model{Model: mdl.ModelSlug, Year: int(mdl.Year), DeviceDefinitionID: mdl.ID}
		}
	}

	return devices, nil

}
