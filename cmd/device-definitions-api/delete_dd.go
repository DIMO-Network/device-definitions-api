package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type deleteDefinition struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*deleteDefinition) Name() string { return "delete-dd" }
func (*deleteDefinition) Synopsis() string {
	return "delete device definitions"
}
func (*deleteDefinition) Usage() string {
	return `delete-dd <id>`
}

func (p *deleteDefinition) SetFlags(_ *flag.FlagSet) {
}

func (p *deleteDefinition) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Prompt the user for input
	fmt.Print("Enter Manufacturer Name: ")

	// Create a new reader
	reader := bufio.NewReader(os.Stdin)

	// Read input from the user
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return subcommands.ExitSuccess
	}
	// Trim whitespace (e.g., newline characters)
	manufacturer := strings.TrimSpace(input)

	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	deviceDefinitionCatalogService := gateways.NewDeviceDefinitionCatalogService(&p.settings, &p.logger)

	id := os.Args[len(os.Args)-1]

	deleted, err := deviceDefinitionCatalogService.Delete(ctx, manufacturer, id)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to delete.")
	}
	// The worker delete is synchronous; nothing to poll.
	fmt.Println("Deleted device definition: ", *deleted)

	return subcommands.ExitSuccess
}
