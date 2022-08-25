package main

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

func migrateDatabase(ctx context.Context, logger zerolog.Logger, s *config.Settings, args []string) {
	command := "up"
	if len(args) > 2 {
		command = args[2]
		if command == "down-to" || command == "up-to" {
			command = command + " " + args[3]
		}
	}

	sqlDb := db.NewDbConnectionFromSettings(ctx, s, true)
	sqlDb.WaitForDB(logger)

	if command == "" {
		command = "up"
	}

	_, err := sqlDb.DBS().Writer.Exec("CREATE SCHEMA IF NOT EXISTS device_definitions_api;")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not create schema:")
	}
	goose.SetTableName("device_definitions_api.migrations")
	if err := goose.Run(command, sqlDb.DBS().Writer.DB, "internal/infrastructure/db/migrations"); err != nil {
		logger.Fatal().Err(err).Msg("failed to apply go code migrations")
	}
}
