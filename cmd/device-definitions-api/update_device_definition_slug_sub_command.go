package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type updateDeviceDefinitionSlugCmd struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*updateDeviceDefinitionSlugCmd) Name() string { return "update-slug" }
func (*updateDeviceDefinitionSlugCmd) Synopsis() string {
	return "update slug in device definition table"
}
func (*updateDeviceDefinitionSlugCmd) Usage() string {
	return `update-slug`
}

func (p *updateDeviceDefinitionSlugCmd) SetFlags(_ *flag.FlagSet) {
}

func (p *updateDeviceDefinitionSlugCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

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
	fmt.Printf("Found %d device-definition(s) in all device-types\n", len(all))

	counter := 1
	slugsUpdatedCounter := 0
	for _, dd := range all {
		slugMMY := common.DeviceDefinitionSlug(dd.R.DeviceMake.NameSlug, dd.ModelSlug, dd.Year)
		if !strings.EqualFold(dd.NameSlug, slugMMY) {
			fmt.Printf("DD => %s updating slug: %s => %s \n", dd.ID, dd.NameSlug, slugMMY)
			dd.NameSlug = slugMMY
			_, err = dd.Update(ctx, pdb.DBS().Writer, boil.Whitelist(models.DeviceDefinitionColumns.NameSlug))
			if err != nil {
				p.logger.Error().Err(err).Send()
			}
			slugsUpdatedCounter++
		}
		if counter%100 == 0 {
			fmt.Printf("DD's processed: %d\n", counter)
		}
		counter++
	}

	fmt.Printf("success. Updated %d slugs\n", slugsUpdatedCounter)
	return subcommands.ExitSuccess
}
