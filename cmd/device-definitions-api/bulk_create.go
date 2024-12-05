package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
		if record[0] == "DefinitionID" {
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
			// read from database, check that it exists there
			dbDefinition, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.NameSlug.EQ(record[0]),
				qm.Load(models.DeviceDefinitionRels.DeviceMake)).One(ctx, pdb.DBS().Reader)
			if err != nil {
				fmt.Println("Error getting definition from DB: ", record[0], err)
			}
			deviceMake := &models.DeviceMake{}
			if dbDefinition != nil {
				deviceMake = dbDefinition.R.DeviceMake
				// go ahead and create on chain
				fmt.Println("trying to create new on chain dd for ", record[0])
				trx, err := deviceDefinitionOnChainService.Create(ctx, *dbDefinition.R.DeviceMake, *dbDefinition)
				if err != nil {
					fmt.Println("Error creating device: ", record[0], err)
					continue
				}
				if trx == nil {
					fmt.Println("transaction null, stopping: ", record[0])
					break
				}
				dbDefinition.TRXHashHex = append(dbDefinition.TRXHashHex, *trx)
				fmt.Println("Created device: ", record[0], *trx)

				// check for trx status
				if len(*trx) > 0 {
					fmt.Println("Created definition: ", record[0], " with transaction hash: ", *trx)
					trxFinished := false
					loops := 0
					for !trxFinished {
						loops++
						time.Sleep(time.Second * 2)
						trxFinished, err = checkTransactionStatus(*trx, p.settings.PolygonScanAPIKey)
						if err != nil {
							fmt.Println("Error checking transaction status: ", err)
						}
						fmt.Println("Transaction status: ", trxFinished)
						if loops > 10 {
							// get device definition from on chain to see if maybe got created but trx still showing false
							onchainDD, err := deviceDefinitionOnChainService.GetDefinitionByID(ctx, record[0], pdb.DBS().Reader)
							fmt.Println("onchainDD: ", onchainDD, err)
							if onchainDD != nil {
								break
							}
						}
					}
				} else {
					fmt.Println("---------no new trx for: ", record[0])
				}
			} else {
				fmt.Println("No device definition found for: ", record[0])
				manufacturerSlug := split[0]
				deviceMake, err = models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(manufacturerSlug)).One(ctx, pdb.DBS().Reader)
				if err != nil {
					fmt.Println("Error getting manufacturer: ", manufacturerSlug, err)
					continue
				}
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
			// repository create
			hardwareTemplateID := "130"
			result, err := definitionRepository.GetOrCreate(ctx, nil, "manual", "", deviceMake.ID, modelName, year, common.DefaultDeviceType, null.JSON{}, true,
				&hardwareTemplateID)
			if err != nil {
				fmt.Println("Error creating definition: ", record[0], err)
				continue
			}
			if len(result.TRXHashHex) > 0 {
				fmt.Println("Created definition: ", record[0], " with transaction hash: ", result.TRXHashHex)
				trxFinished := false
				loops := 0
				for !trxFinished {
					loops++
					time.Sleep(time.Second * 2)
					trxFinished, err = checkTransactionStatus(result.TRXHashHex[0], p.settings.PolygonScanAPIKey)
					if err != nil {
						fmt.Println("Error checking transaction status: ", err)
					}
					fmt.Println("Transaction status: ", trxFinished)
					if loops > 10 {
						// get device definition from on chain to see if maybe got created but trx still showing false
						onchainDD, err := deviceDefinitionOnChainService.GetDefinitionByID(ctx, record[0], pdb.DBS().Reader)
						fmt.Println("onchainDD: ", onchainDD, err)
						if onchainDD != nil {
							break
						}
					}
				}
			} else {
				fmt.Println("---------no new trx for: ", record[0], "updated at: ", result.UpdatedAt.String())
			}

		}

	}

	return subcommands.ExitSuccess
}

func checkTransactionStatus(txHash, apiKey string) (bool, error) {
	url := fmt.Sprintf("https://api.polygonscan.com/api?module=transaction&action=gettxreceiptstatus&txhash=%s&apikey=%s", txHash, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var txStatus TxStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&txStatus); err != nil {
		return false, err
	}

	// Check the transaction status
	if txStatus.Status == "1" && txStatus.Result.Status == "1" {
		return true, nil
	}
	return false, nil
}

type TxStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Status string `json:"status"`
	} `json:"result"`
}
