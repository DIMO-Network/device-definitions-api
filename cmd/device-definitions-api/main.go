package main

import (
	"log"
	"os"

	_ "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/docs"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/api"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
)

func main() {

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
		migrateDatabase(settings, os.Args)
	default:
		api.Run(settings)
	}
}
