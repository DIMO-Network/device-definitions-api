package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/shared/db"
	"github.com/pressly/goose/v3"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type migrateDBCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	up   bool
	down bool
}

func (*migrateDBCmd) Name() string     { return "migrate" }
func (*migrateDBCmd) Synopsis() string { return "migrate args to stdout." }
func (*migrateDBCmd) Usage() string {
	return `migrate [-up-to|-down-to] <some text>:
	migrate args.
  `
}

func (p *migrateDBCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.up, "up", false, "up database")
	f.BoolVar(&p.down, "down", false, "down database")
}

func (p *migrateDBCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	command := "up"
	if p.down {
		command = "down"
	}

	sqlDb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	sqlDb.WaitForDB(p.logger)

	_, err := sqlDb.DBS().Writer.Exec("CREATE SCHEMA IF NOT EXISTS device_definitions_api;")
	if err != nil {
		p.logger.Fatal().Err(err).Msg("could not create schema:")
	}
	goose.SetTableName("device_definitions_api.migrations")
	if err := goose.Run(command, sqlDb.DBS().Writer.DB, "internal/infrastructure/db/migrations"); err != nil {
		p.logger.Fatal().Err(err).Msg("failed to apply go code migrations")
	}
	return subcommands.ExitSuccess
}
