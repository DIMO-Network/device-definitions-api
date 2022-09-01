package main

import (
	"context"
	"log"
	"os"

	_ "github.com/DIMO-Network/device-definitions-api/docs"
	"github.com/DIMO-Network/device-definitions-api/internal/api"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "device-definitions-api").
		Str("git-sha1", gitSha1).
		Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		log.Fatal("could not load settings: $s", err)
	}

	switch arg {
	case "migrate":
		migrateDatabase(ctx, logger, &settings, os.Args)
	case "search-sync-dds":
		searchSyncDD(ctx, &settings, logger)
	case "ipfs-sync-data":
		ipfsSyncData(ctx, &settings, logger)
	case "smartcar-compatibility":
		smartCarCompatibility(ctx, &settings, logger)
	case "smartcar-sync":
		smartCarSync(ctx, &settings, logger)
	default:
		api.Run(ctx, logger, &settings)
	}
}
