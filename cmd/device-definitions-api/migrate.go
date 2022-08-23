package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
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

	totalTime := 0
	for !sqlDb.IsReady() {
		if totalTime > 30 {
			fmt.Println("could not connect to postgres after 30 seconds")
		}
		time.Sleep(time.Second)
		totalTime++
	}

	if command == "" {
		command = "up"
	}
	if !sqlDb.IsReady() {
		time.Sleep(1 * time.Second)
		fmt.Println("db not ready, retrying in 1s")
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
