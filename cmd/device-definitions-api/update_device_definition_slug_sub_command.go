package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared"
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

	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		models.DeviceDefinitionWhere.Year.GTE(2012),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, pdb.DBS().Reader)

	if err != nil {
		p.logger.Error().Err(err).Send()
	}

	for _, dd := range all {
		dd.NameSlug = shared.SlugString(fmt.Sprintf("%s_%s_%d", dd.R.DeviceMake.NameSlug, dd.ModelSlug, dd.Year))
		if err = dd.Upsert(ctx, pdb.DBS().Writer, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
			p.logger.Error().Err(err).Send()
		}
		p.logger.Info().Msgf("DD => %s updated slug => %s", dd.ID, dd.NameSlug)
	}

	fmt.Printf("success")
	return subcommands.ExitSuccess
}
