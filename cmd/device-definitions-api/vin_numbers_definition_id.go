package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// can delete this whole file soon
type setVinNumbersDefinitionID struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*setVinNumbersDefinitionID) Name() string { return "vin-numbers-definition-id" }
func (*setVinNumbersDefinitionID) Synopsis() string {
	return "set definition-id in vin_numbers"
}
func (*setVinNumbersDefinitionID) Usage() string {
	return `vin-numbers-definition-id`
}

func (p *setVinNumbersDefinitionID) SetFlags(_ *flag.FlagSet) {
}

func (p *setVinNumbersDefinitionID) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	fmt.Printf("Starting processing vin numbers\n")

	all, err := models.VinNumbers(models.VinNumberWhere.DefinitionID.IsNull()).All(ctx, pdb.DBS().Reader)

	if err != nil {
		p.logger.Error().Err(err).Send()
		return subcommands.ExitFailure
	}
	fmt.Printf("Found %d vin numbers\n", len(all))

	counter := 1
	slugsUpdatedCounter := 0
	for _, vn := range all {
		definition, err := models.FindDeviceDefinition(ctx, pdb.DBS().Reader, vn.DeviceDefinitionID)
		if err != nil {
			p.logger.Error().Err(err).Msgf("Error finding device definition %s", vn.DeviceDefinitionID)
			continue
		}
		vn.DefinitionID = null.StringFrom(definition.NameSlug)
		counter++
		_, err = vn.Update(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			p.logger.Fatal().Err(err).Send()
		}
	}

	fmt.Printf("success. set %d slugs\n", slugsUpdatedCounter)
	return subcommands.ExitSuccess
}
