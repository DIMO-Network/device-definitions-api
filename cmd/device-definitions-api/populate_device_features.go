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

type SupportLevelEnum int8

const (
	NotSupported   SupportLevelEnum = 0
	MaybeSupported SupportLevelEnum = 1
	Supported      SupportLevelEnum = 2
)

func (r SupportLevelEnum) Int() int8 {
	return int8(r)
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
				logger.Err(err).Str("integrationId", intID).Str("deviceDefinitionId", ddID).Msg("error occurred fetching device")
				continue
			}

			/* if len(devices) > 1 { // we have for multiple continents
				// handle when we have region in elasticsearch
			} */

			if len(devices) < 1 {
				// handle not found
				logger.Err(err).Str("integrationId", intID).Str("deviceDefinitionId", ddID).Msg("error could not find device")
				continue
			}

			feature := prepareFeatureData(d.Features.Buckets)

			device := devices[0]
			err = device.Features.Marshal(&feature)
			if err != nil {
				logger.Err(err).Str("integrationId", intID).Str("deviceDefinitionId", ddID).Msg("could not marshal feature information into device")
				continue
			}

			if _, err := device.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				logger.Err(err).Str("integrationId", intID).Str("deviceDefinitionId", ddID).Msg("could not update device")
				continue
			}
		}
	}

	return nil
}

func prepareFeatureData(f map[string]map[string]int) []elasticModels.DeviceIntegrationFeatures {
	ft := []elasticModels.DeviceIntegrationFeatures{}

	for k, v := range f {
		var supportLevel int8

		if v["doc_count"] > 0 {
			supportLevel = int8(Supported)
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
