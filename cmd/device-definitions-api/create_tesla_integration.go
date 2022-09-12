package main

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var teslaRegions = []string{common.AmericasRegion.String(), common.EuropeRegion.String()}

// createTeslaIntegrations ensures that we have a Tesla integration and that it is attached to all
// Tesla device definitions in our supported countries. This behaves well if some of these records
// already exist.
func createTeslaIntegrations(ctx context.Context, logger zerolog.Logger, settings *config.Settings) error {
	pdb := db.NewDbConnectionFromSettings(ctx, settings, true)
	pdb.WaitForDB(logger)

	tx, err := pdb.DBS().Writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint

	var teslaInt *models.Integration

	// Check to see if a Tesla integration exists that we can use. If there is none, create one.
	if teslaInts, err := models.Integrations(models.IntegrationWhere.Vendor.EQ("Tesla")).All(ctx, tx); err != nil {
		return fmt.Errorf("failed searching for existing Tesla integrations: %w", err)
	} else if len(teslaInts) > 1 {
		return fmt.Errorf("found %d > 1 existing Tesla integrations, unclear which to use", len(teslaInts))
	} else if len(teslaInts) == 1 {
		teslaInt = teslaInts[0]
		logger.Info().Msgf("Found an existing Tesla integration with id %s", teslaInt.ID)
	} else {
		teslaInt = &models.Integration{
			ID:     ksuid.New().String(),
			Vendor: "Tesla",
			Type:   models.IntegrationTypeAPI,
			Style:  models.IntegrationStyleOEM,
		}
		if err := teslaInt.Insert(ctx, tx, boil.Infer()); err != nil {
			return fmt.Errorf("failed to create Tesla integration: %w", err)
		}
		logger.Info().Msgf("Created new Tesla integration with id %s", teslaInt.ID)
	}

	// Grab all Tesla device definitions, along with any existing Tesla integration links. It would
	// be nice to only load definitions that are missing the integration, but the SQLBoiler is a
	// bit awkward.
	teslaMake, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ("Tesla")).One(ctx, pdb.DBS().Reader)
	if err != nil {
		return errors.Wrap(err, "could not retrieve Tesla make, check it exists")
	}
	teslaDefs, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.DeviceMakeID.EQ(teslaMake.ID),
		qm.Load(
			models.DeviceDefinitionRels.DeviceIntegrations,
			models.DeviceIntegrationWhere.IntegrationID.EQ(teslaInt.ID),
		),
	).All(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to look up all Tesla device definitions: %w", err)
	}

	for _, teslaDef := range teslaDefs {
		integRegions := shared.NewStringSet()
		for _, integ := range teslaDef.R.DeviceIntegrations {
			integRegions.Add(integ.Region)
		}

		for _, region := range teslaRegions {
			if !integRegions.Contains(region) {
				integ := &models.DeviceIntegration{
					DeviceDefinitionID: teslaDef.ID,
					IntegrationID:      teslaInt.ID,
					Region:             region,
				}
				if err := integ.Insert(ctx, tx, boil.Infer()); err != nil {
					return fmt.Errorf("failed to link integration with device definition %s in region %s: %w", teslaDef.ID, region, err)
				}
				logger.Info().Msgf("Created integration for %d %s in %s", teslaDef.Year, teslaDef.Model, region)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit Tesla integrations: %w", err)
	}

	return nil
}
