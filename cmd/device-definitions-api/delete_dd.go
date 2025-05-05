package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	dd_common "github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ethereum/go-ethereum/ethclient"
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
	send, err := createSender(ctx, &p.settings, &p.logger)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create sender.")
	}

	ethClient, err := ethclient.Dial(p.settings.EthereumRPCURL.String())
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(&p.settings, &p.logger, ethClient, chainID, send, pdb.DBS)

	id := os.Args[len(os.Args)-1]

	trx, err := deviceDefinitionOnChainService.Delete(ctx, manufacturer, id)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to delete.")
	}

	if len(*trx) > 0 {
		trxFinished := false
		loops := 0
		for !trxFinished {
			loops++
			time.Sleep(time.Second * 2)
			trxFinished, err = dd_common.CheckTransactionStatus(*trx, p.settings.PolygonScanAPIKey, !p.settings.IsProd())
			if err != nil {
				fmt.Println("Error checking transaction status: ", err)
			}
			fmt.Println("Transaction status: ", trxFinished)
			if loops > 10 {
				// get device definition from on chain to see if maybe got created but trx still showing false
				onchainDD, _, err := deviceDefinitionOnChainService.GetDefinitionByID(ctx, id, pdb.DBS().Reader)
				fmt.Println("onchainDD: ", onchainDD, err)
				if onchainDD != nil {
					break
				}
			}
		}
	}

	return subcommands.ExitSuccess
}
