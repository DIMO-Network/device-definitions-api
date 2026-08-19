package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/api"

	_ "github.com/DIMO-Network/device-definitions-api/docs"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared/pkg/settings"
	"github.com/rs/zerolog"
)

// @title                      DIMO Device Definitions API
// @version                    1.0
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()

	settings, err := settings.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		log.Fatal("could not load settings: $s", err)
	}
	level, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		log.Fatal("could not parse log level: $s", err)
	}
	logger := zerolog.New(os.Stdout).Level(level).With().
		Timestamp().
		Str("app", settings.ServiceName).
		Str("git-sha1", gitSha1).
		Logger()
	identity := gateways.NewIdentityAPIService(&logger, &settings)

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&migrateDBCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&addVINCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&addVINsCSVCmd{logger: logger, settings: settings, identity: identity}, "")
	subcommands.Register(&decodeVINCmd{logger: &logger, settings: &settings}, "")
	subcommands.Register(&syncDeviceDefinitionSearchCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&deleteDefinition{logger: logger, settings: settings}, "")
	subcommands.Register(&bulkUpdatePowertrain{logger: logger, settings: settings}, "")

	if len(os.Args) == 1 {
		// Run API & everythying else
		api.Run(ctx, logger, &settings)
	} else {
		flag.Parse()
		os.Exit(int(subcommands.Execute(ctx)))
	}

}
