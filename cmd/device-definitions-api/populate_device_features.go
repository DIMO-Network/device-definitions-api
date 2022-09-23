package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	elastic "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch"
	elasticModels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type jsonObj map[string]interface{}

func prepareFeatureData(f map[string]map[string]int) []elasticModels.DeviceIntegrationFeatures {
	ft := []elasticModels.DeviceIntegrationFeatures{}

	for k, v := range f {
		supportLevel := 0

		if v["doc_count"] > 0 {
			supportLevel = 2
		}

		feat := elasticModels.DeviceIntegrationFeatures{
			FeatureKey:   k,
			SupportLevel: int8(supportLevel),
		}

		ft = append(ft, feat)
	}

	return ft
}

func getIntegrationFeatures(ctx context.Context, d db.Store) (string, error) {
	ifeats, err := models.IntegrationFeatures().All(ctx, d.DBS().Reader)
	if err != nil {
		return "", err
	}

	filters := jsonObj{}

	for _, v := range ifeats {
		esKey := v.ElasticProperty
		if v.FeatureKey == "tires" {
			esKey = v.FeatureKey
		}
		filters[esKey] = jsonObj{"exists": jsonObj{"field": "data." + v.ElasticProperty}}
	}

	esFilters, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}

	return string(esFilters), nil
}

func populateDeviceFeaturesFromEs(ctx context.Context, logger zerolog.Logger, s *config.Settings) error {
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	es, _ := elastic.NewElasticSearch(s, logger)

	esFilters, err := getIntegrationFeatures(ctx, pdb)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load integration features")
	}

	resp, err := es.GetDeviceFeatures(s.Environment, esFilters)
	if err != nil {
		return err
	}

	for _, f := range resp.Aggregations.Features.Buckets {
		intID := strings.TrimPrefix(f.Key, "dimo/integration/")

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

			/* if len(devices) > 1 { // we have for multiple continents
				// handle when we have region in elasticsearch
			} */

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
