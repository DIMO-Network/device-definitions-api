package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

type bulkCreateDefinitions struct {
	logger   zerolog.Logger
	settings config.Settings
}

func (*bulkCreateDefinitions) Name() string { return "bulk-create" }
func (*bulkCreateDefinitions) Synopsis() string {
	return "bulk create device definitions"
}
func (*bulkCreateDefinitions) Usage() string {
	return `bulk-create`
}

func (p *bulkCreateDefinitions) SetFlags(_ *flag.FlagSet) {
}

func (p *bulkCreateDefinitions) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(&p.settings, &p.logger, ethClient, chainID, send)

	definitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, deviceDefinitionOnChainService)

	fmt.Printf("Starting processing bulk definitions to create\n")
	// open a csv file in the tmp directory
	file, err := os.Open("/tmp/definitions.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return subcommands.ExitFailure
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all the records from the CSV
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return subcommands.ExitFailure
	}

	inputReader := bufio.NewReader(os.Stdin) // used for cli prompt
	mappedModels := map[string]string{}      // keep track of mapped models

	// read through each line, forach:
	for _, record := range records {
		fmt.Println("-----------", record[0])
		if record[0] == "DefinitionId" {
			continue // skip first row header
		}
		dd, err := deviceDefinitionOnChainService.GetDefinitionByID(ctx, record[0], pdb.DBS().Reader)
		if err != nil {
			fmt.Println("Error getting definition: ", record[0], err)
		}
		if dd == nil {
			fmt.Println("Definition not found, will try to create it...: ", record[0])
			split := strings.Split(record[0], "_")
			if len(split) != 3 {
				continue
			}
			manufacturerSlug := split[0]
			deviceMake, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(manufacturerSlug)).One(ctx, pdb.DBS().Reader)
			if err != nil {
				fmt.Println("Error getting manufacturer: ", manufacturerSlug, err)
				continue
			}
			modelSlug := split[1]
			year, yrErr := strconv.Atoi(split[2])
			if yrErr != nil {
				fmt.Println("Error getting year: ", split[2], err)
				continue
			}
			modelName := ""
			// if slug exists in map can short circuit below
			if mappedModels[modelSlug] != "" {
				modelName = mappedModels[modelSlug]
			}
			if modelName == "" {
				// determine the model
				deviceDefinition, _ := models.DeviceDefinitions(models.DeviceDefinitionWhere.ModelSlug.EQ(modelSlug)).One(ctx, pdb.DBS().Reader)
				if deviceDefinition != nil {
					modelName = deviceDefinition.Model
				} else {
					// prompt in cli
					// Read input until the user hits Enter
					fmt.Printf("Enter model name for slug %s : ", modelSlug)
					input, err := inputReader.ReadString('\n')
					if err != nil {
						fmt.Println("Error reading input:", err)
						return subcommands.ExitFailure
					}
					input = strings.TrimSpace(input)
					modelName = input
				}
				mappedModels[modelSlug] = modelName
			}
			// repository create?
			hardwareTemplateID := "130"
			_, err = definitionRepository.GetOrCreate(ctx, nil, "ruptela", "", deviceMake.ID, modelName, year, common.DefaultDeviceType, null.JSON{}, true,
				&hardwareTemplateID)
			if err != nil {
				fmt.Println("Error creating definition: ", record[0], err)
				continue
			}
			fmt.Println("---------Created definition: ", record[0])
		}

	}

	return subcommands.ExitSuccess
}
