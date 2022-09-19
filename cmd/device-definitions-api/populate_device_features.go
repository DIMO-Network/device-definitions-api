package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/pkg/elastic"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
)

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
			// features := d.Features.Buckets

			devices, err := models.DeviceIntegrations(
				models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(ddID),
				models.DeviceIntegrationWhere.IntegrationID.EQ(intID),
			).All(ctx, pdb.DBS().Reader)
			// check if device exists
			if err != nil {
				return fmt.Errorf("error occurred fetching device with integration id %s and deviceDefinitionId %s. error: %s", intID, ddID, err.Error())
			}

			if len(devices) > 1 { // we have for multiple continents
				// handle when we have region in elasticsearch
			}

			if len(devices) < 1 {
				// handle not found
				return fmt.Errorf("error could not find device with integration id %s and deviceDefinitionId %s.", intID, ddID)
			}

			/* esFt := d.Features.Buckets

			device := devices[0]
			feat := models.DeviceIntegrationFeatures{

			} */
			logger.Info().Str("integrationID", intID).Str("ddID", ddID).Msg("Data from es")
		}
	}

	return nil
}
