package main

import (
	"context"
	"flag"
	"os"

	"github.com/DIMO-Network/shared/db"
	"github.com/pressly/goose/v3"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type databaseCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	up   bool
	down bool
}

func (*databaseCmd) Name() string     { return "migrate" }
func (*databaseCmd) Synopsis() string { return "migrate args to stdout." }
func (*databaseCmd) Usage() string {
	return `migrate [-up-to|-down-to] <some text>:
	migrate args.
  `
}

func (p *databaseCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.up, "up", false, "up database")
	f.BoolVar(&p.down, "down", false, "down database")
}

func (p *databaseCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	command := "up"
	if len(f.Args()) > 2 {
		command = f.Args()[2]
		if p.down || p.up {
			command = command + " " + os.Args[3] // migration name
		}
	}

	sqlDb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	sqlDb.WaitForDB(p.logger)

	if command == "" {
		command = "up"
	}

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
