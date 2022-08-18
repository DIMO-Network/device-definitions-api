package main

import (
	"context"
	"log"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/config"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func migrateDatabase(ctx context.Context, s *config.Settings, args []string) {
	command := "up"
	if len(args) > 2 {
		command = args[2]
		if command == "down-to" || command == "up-to" {
			command = command + " " + args[3]
		}
	}

	sqlDb := db.NewDbConnectionFromSettings(ctx, s, true)

	if command == "" {
		command = "up"
	}

	_, err := sqlDb.DBS().Writer.Exec("CREATE SCHEMA IF NOT EXISTS device_definitions_api;")
	if err != nil {
		log.Fatal("could not create schema: $s", err)
	}
	goose.SetTableName("device_definitions_api.migrations")
	if err := goose.Run(command, sqlDb.DBS().Writer.DB, "internal/infrastructure/db/migrations"); err != nil {
		log.Fatal("failed to apply go code migrations: $s", err)
	}
}
