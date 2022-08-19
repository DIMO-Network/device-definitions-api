package main

import (
	"context"
	"log"
	"os"

	_ "github.com/DIMO-Network/device-definitions-api/docs"
	"github.com/DIMO-Network/device-definitions-api/internal/api"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
)

func main() {
	ctx := context.Background()
	arg := ""
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		log.Fatal("could not load settings: $s", err)
	}

	switch arg {
	case "migrate":
		migrateDatabase(ctx, &settings, os.Args)
	default:
		api.Run(ctx, &settings)
	}
}
