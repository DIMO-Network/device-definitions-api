package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/api"

	_ "github.com/DIMO-Network/device-definitions-api/docs"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
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
	subcommands.Register(&syncFuelImageCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&addVINCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&powerTrainTypeCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&decodeVINCmd{logger: &logger, settings: &settings}, "")
	subcommands.Register(&syncDeviceDefinitionSearchCmd{logger: logger, settings: settings}, "")
	subcommands.Register(&deleteDefinition{logger: logger, settings: settings}, "")
	subcommands.Register(&syncR1CompatibiltyCmd{logger: logger, settings: settings}, "")

	if len(os.Args) == 1 {
		// Run API & everythying else
		sigSender, err := createSender(ctx, &settings, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create sender.")
		}
		api.Run(ctx, logger, &settings, sigSender)
	} else {
		flag.Parse()
		os.Exit(int(subcommands.Execute(ctx)))
	}

}

func createSender(ctx context.Context, settings *config.Settings, logger *zerolog.Logger) (sender.Sender, error) {
	if settings.PrivateKeyMode {
		logger.Warn().Msg("Using injected private key. Never do this in production.")
		send, err := sender.FromKey(settings.SenderPrivateKey)
		if err != nil {
			return nil, err
		}
		logger.Info().Str("address", send.Address().Hex()).Msg("Loaded private key account.")
		return send, nil
	}

	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	kmsc := kms.NewFromConfig(awsconf)
	send, err := sender.FromKMS(ctx, kmsc, settings.KMSKeyID)
	if err != nil {
		return nil, err
	}
	logger.Info().Msgf("Loaded KMS key %s, address %s.", settings.KMSKeyID, send.Address().Hex())
	return send, nil
}
