package main

import (
	"context"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/pkg/elastic"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func prepareFeatureData(f map[string]map[string]int) []models.DeviceIntegrationFeatures {
	displayNameMapping := map[string]string{
		"batteryVoltage":       "Battery Voltage",
		"fuelPercentRemaining": "Fuel Tank",
		"odometer":             "Odometer",
		"oil":                  "Engine Oil Life",
		"soc":                  "EV Battery",
		"tires":                "Tires",
		"speed":                "Speed",
	}

	var ft []models.DeviceIntegrationFeatures

	for k, v := range f {
		feat := models.DeviceIntegrationFeatures{}
		feat.ElasticProperty = k
		feat.DisplayName = displayNameMapping[k]
		feat.SupportLevel = 0

		if v["doc_count"] > 0 {
			feat.SupportLevel = 2
		}

		ft = append(ft, feat)
	}

	return ft
}

func populateDeviceFeaturesFromEs(ctx context.Context, logger zerolog.Logger, s *config.Settings) error {
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	es, _ := elastic.NewElasticSearhBaseService(s, logger)

	resp, err := es.GetDeviceFeatures(s.Environment)
	if err != nil {
		return err
	}

	for _, f := range resp.Aggregations.Features.Buckets {
		intID := strings.Split(f.Key, "/")[2]

		for _, d := range f.DeviceDefinitions.Buckets {
			ddID := d.Key

			devices, err := models.DeviceIntegrations(
				models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(ddID),
				models.DeviceIntegrationWhere.IntegrationID.EQ(intID),
			).All(ctx, pdb.DBS().Reader)
			// check if device exists
			if err != nil {
				logger.Error().Msgf("error occurred fetching device with integration id %s and deviceDefinitionId %s. error: %s", intID, ddID, err.Error())
				continue
			}

			if len(devices) > 1 { // we have for multiple continents
				// handle when we have region in elasticsearch
			}

			if len(devices) < 1 {
				// handle not found
				logger.Error().Msgf("error could not find device with integration id %s and deviceDefinitionId %s", intID, ddID)
				continue
			}

			feature := prepareFeatureData(d.Features.Buckets)

			device := devices[0]

			err = device.Features.Marshal(&feature)
			if err != nil {
				logger.Error().Msgf("could not marshal feature information into device with integration id %s and deviceDefinitionId %s:", intID, ddID)
				continue
			}

			if _, err := device.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				logger.Error().Msgf("could not update device with integration id %s and deviceDefinitionId %s:", intID, ddID)
				continue
			}
		}
	}

	return nil
}
