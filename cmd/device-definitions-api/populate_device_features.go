package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	elastic "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch"
	elasticModels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type jsonObj map[string]any

type SupportLevelEnum int8

const (
	NotSupported   SupportLevelEnum = 0
	MaybeSupported SupportLevelEnum = 1 //nolint
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

	for _, i := range resp.Aggregations.Integrations.Buckets {
		intID := strings.TrimPrefix(i.Key, "dimo/integration/")

		integration, err := models.FindIntegration(ctx, pdb.DBS().Reader, intID)
		if err != nil {
			logger.Err(err).Msg("Error occurred fetching integration.")
			continue
		}

		for _, d := range i.DeviceDefinitions.Buckets {
			ddID := d.Key

			deviceDef, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.EQ(ddID),
				qm.Load(models.DeviceDefinitionRels.DeviceMake)).One(ctx, pdb.DBS().Reader)
			if err != nil {
				logger.Err(err).Msg("Error occurred fetching device definition.")
				continue
			}
			// skip smartcar integration if Tesla
			if integration.Vendor == common.SmartCarVendor && deviceDef.R.DeviceMake.NameSlug == "tesla" {
				continue
			}

			for _, r := range d.Regions.Buckets {
				region := r.Key
				logger := logger.With().Str("integrationId", intID).Str("deviceDefinitionId", ddID).Str("region", region).Logger()

				deviceInt, err := models.FindDeviceIntegration(ctx, pdb.DBS().Reader, ddID, intID, region)
				insert := false
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						insert = true
						deviceInt = &models.DeviceIntegration{
							DeviceDefinitionID: ddID,
							IntegrationID:      intID,
							Region:             region,
						}
					} else {
						logger.Err(err).Msg("Eror occurred fetching device integration.")
						continue
					}
				}

				feature := prepareFeatureData(r.Features.Buckets, deviceDef)

				err = deviceInt.Features.Marshal(&feature)
				if err != nil {
					logger.Err(err).Msg("could not marshal feature information into device integration.")
					continue
				}
				if insert {
					err = deviceInt.Insert(ctx, pdb.DBS().Writer, boil.Infer())
					if err != nil {
						logger.Err(err).Msg("could not insert device integration with feature information.")
					}
				} else {
					if _, err := deviceInt.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
						logger.Err(err).Msg("could not update device integration with feature information.")
					}
				}
			}
		}
	}
	logger.Info().Msgf("processed %d integrations from elastic", len(resp.Aggregations.Integrations.Buckets))

	return nil
}

func prepareFeatureData(i map[string]elastic.ElasticFilterResult, def *models.DeviceDefinition) []elasticModels.DeviceIntegrationFeatures {
	ft := []elasticModels.DeviceIntegrationFeatures{}

	for k, v := range i {
		supportLevel := NotSupported.Int()

		if v.DocCount > 0 {
			supportLevel = Supported.Int()
		}
		// manual override for VIN support
		if k == "vin" && def.R.DeviceMake != nil {
			if def.Year >= 2011 && def.R.DeviceMake.NameSlug != "skoda" { // exclude skoda hardcode
				supportLevel = Supported.Int()
			} else if def.Year >= 2006 && def.R.DeviceMake.NameSlug == "mercedes-benz" { // include mercedes hard code
				supportLevel = Supported.Int()
			}
		}

		feat := elasticModels.DeviceIntegrationFeatures{
			FeatureKey:   k,
			SupportLevel: supportLevel,
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
		filters[v.FeatureKey] = jsonObj{
			"exists": jsonObj{
				"field": "data." + v.ElasticProperty,
			},
		}
	}

	esFilters, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}

	return string(esFilters), nil
}
