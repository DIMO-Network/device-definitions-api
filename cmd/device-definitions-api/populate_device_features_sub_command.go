package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"strconv"
	"strings"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
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

type syncDeviceFeatureCmd struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*syncDeviceFeatureCmd) Name() string     { return "populate-device-features" }
func (*syncDeviceFeatureCmd) Synopsis() string { return "populate-device-features args to stdout." }
func (*syncDeviceFeatureCmd) Usage() string {
	return `populate-device-features [] <some text>:
	sync args.
  `
}

func (p *syncDeviceFeatureCmd) SetFlags(_ *flag.FlagSet) {

}

func (p *syncDeviceFeatureCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := populateDeviceFeaturesFromEs(ctx, p.logger, &p.settings)
	if err != nil {
		p.logger.Error().Err(err)
	}
	return subcommands.ExitSuccess
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
				qm.Load(models.DeviceDefinitionRels.DeviceMake),
				qm.Load(models.DeviceDefinitionRels.DeviceType)).One(ctx, pdb.DBS().Reader)
			if err != nil {
				logger.Err(err).Msg("Error occurred fetching device definition.")
				continue
			}
			// skip smartcar integration if Tesla
			if integration.Vendor == common.SmartCarVendor && deviceDef.R.DeviceMake.NameSlug == "tesla" {
				continue
			}
			// map of regions and features for this dd, then fill in
			regionToFeatures := map[string][]elasticModels.DeviceIntegrationFeatures{}

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
						logger.Err(err).Msgf("error occurred fetching device integration dd_id %s", ddID)
						continue
					}
				}

				feature := prepareFeatureData(&logger, r.Features.Buckets, deviceDef)
				// populate the map for future iteration to copy populated region to empty region (autopi only)
				regionToFeatures[region] = feature

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

			if integration.Vendor == common.AutoPiVendor {
				// look at all regions and copy from feature populated ones to ones that have 0 features reported but with SupportLevel as Maybe.
				// note this method also writes to the db for the "other" regions
				copyFeaturesToMissingRegion(ctx, logger, regionToFeatures, pdb, ddID, intID)
			}
		}
	}
	logger.Info().Msgf("processed %d integrations from elastic", len(resp.Aggregations.Integrations.Buckets))

	return nil
}

// prepareFeatureData builds out what the supported features should be based on data from elastic and the device definition
func prepareFeatureData(logger *zerolog.Logger, i map[string]elastic.ElasticFilterResult, def *models.DeviceDefinition) []elasticModels.DeviceIntegrationFeatures {
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
		// manual override for range support when we can calculate it
		if k == "range" && supportLevel == NotSupported.Int() && def.Metadata.Valid {
			// pull out mpg and fuel_tank_capacity_gal to check if can support range
			attrs := common.GetDeviceAttributesTyped(def.Metadata, def.R.DeviceType.Metadatakey)
			var fuelTankCapGal, mpg float64 //mpgHwy
			for _, attr := range attrs {
				switch attr.Name {
				case "fuel_tank_capacity_gal":
					if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
						fuelTankCapGal = v
					}
				case "mpg":
					if v, err := strconv.ParseFloat(attr.Value, 32); err == nil {
						mpg = v
					}
				}
			}
			if fuelTankCapGal > 0 && mpg > 0 {
				logger.Info().Msg("found fuel_tank_capacity_gal and mpg for range calculation")
				// loop over i to check if fuelPercentRemaining exists, if so can support "range"
				for k2, v2 := range i {
					if k2 == "fuelPercentRemaining" || k2 == "fuel_tank" {
						if v2.DocCount > 0 {
							supportLevel = Supported.Int()
							logger.Info().Msgf("range support enable, found signal %s", k2)
						}
					}
				}
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

// copyFeaturesToMissingRegion looks for a region that has no features and tries copying. expects device integrations to exist in the DB but not necessarily in regionToFeatures
func copyFeaturesToMissingRegion(ctx context.Context, logger zerolog.Logger, regionToFeatures map[string][]elasticModels.DeviceIntegrationFeatures, pdb db.Store, ddID, intID string) {
	localLog := logger.With().Str("integration_id", intID).Str("device_definition_id", ddID).Logger()
	all, err := models.DeviceIntegrations(models.DeviceIntegrationWhere.IntegrationID.EQ(intID), models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(ddID)).
		All(ctx, pdb.DBS().Reader)
	if err != nil {
		localLog.Err(err).Msg("error querying device_integrations when copyFeaturesToMissingRegion")
		return
	}
	// populate any non-existent regions for this ddID here
	for _, di := range all {
		if _, ok := regionToFeatures[di.Region]; !ok {
			regionToFeatures[di.Region] = []elasticModels.DeviceIntegrationFeatures{}
		}
	}

	var emptyRegions []string
	populatedRegion := ""
	biggest := 0
	// see if we have both a region with 0 features and a region with many features
	for r, features := range regionToFeatures {
		if len(features) == 0 {
			emptyRegions = append(emptyRegions, r)
		}
		if len(features) > biggest {
			populatedRegion = r
			biggest = len(features)
		}
	}
	// if populated region exists, let's copy over from the populated one to empty regions
	if populatedRegion != "" {
		for _, region := range emptyRegions {
			localLog.Info().Msgf("found a device integration region that has no features. will try copying dd_id %s, %s to %s", ddID, populatedRegion, region)
			deviceInt, err := models.FindDeviceIntegration(ctx, pdb.DBS().Reader, ddID, intID, region)
			if err != nil {
				localLog.Err(err).Msgf("error occurred fetching device integration for empty region %s.", region)
				return
			}
			// set support to 1 on the copy
			features := regionToFeatures[populatedRegion]
			for idxF, f := range features {
				if f.SupportLevel > NotSupported.Int() {
					features[idxF].SupportLevel = MaybeSupported.Int()
				}
			}
			err = deviceInt.Features.Marshal(&features)
			if err != nil {
				localLog.Err(err).Msg("error occurred marshalling feature into device integration")
				return
			}
			if _, err := deviceInt.Update(ctx, pdb.DBS().Writer, boil.Infer()); err != nil {
				localLog.Err(err).Msgf("could not update device integration with feature information region %s", region)
			}
		}
	}
}
