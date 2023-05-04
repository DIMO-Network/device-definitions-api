package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/api"

	_ "github.com/DIMO-Network/device-definitions-api/docs"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

func main() {
	gitSha1 := os.Getenv("GIT_SHA1")
	ctx := context.Background()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
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

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&migrateDBCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&syncOpsCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&syncFuelImageCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&syncDeviceFeatureCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&addVINCmd{logger: logger, settings: settings}, "")

	// Run API
	if len(os.Args) == 1 {
		api.Run(ctx, logger, &settings)
	} else {
		flag.Parse()
		os.Exit(int(subcommands.Execute(ctx)))
	}

}
