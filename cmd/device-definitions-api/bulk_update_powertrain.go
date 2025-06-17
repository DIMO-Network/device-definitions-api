package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type bulkUpdatePowertrain struct {
	logger   zerolog.Logger
	settings config.Settings

	sender sender.Sender
}

func (*bulkUpdatePowertrain) Name() string { return "bulk-update-powertrain" }
func (*bulkUpdatePowertrain) Synopsis() string {
	return "updates definitions from csv file with corresponding definitionId,powertrain pair per row"
}
func (*bulkUpdatePowertrain) Usage() string {
	return `bulk-update-powertrain <filename csv>`
}

func (p *bulkUpdatePowertrain) SetFlags(_ *flag.FlagSet) {
}

func (p *bulkUpdatePowertrain) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	filename := ""
	for _, arg := range os.Args {
		if strings.Contains(arg, ".csv") {
			filename = arg
		}
	}
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return subcommands.ExitFailure
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		return subcommands.ExitFailure
	}

	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	ethClient, err := ethclient.Dial(p.settings.EthereumRPCURL.String())
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}
	onChainSvc := gateways.NewDeviceDefinitionOnChainService(&p.settings, &p.logger, ethClient, chainID, p.sender, pdb.DBS)

	notFoundDefinitions := make([]string, 0)

	for i, record := range records {
		if len(record) < 2 {
			fmt.Printf("Skipping malformed line %d: %v\n", i+1, record)
			continue
		}
		definitionID := strings.ToLower(record[0])
		powertrain := strings.ToUpper(record[1])
		if definitionID == "" || powertrain == "" {
			fmt.Printf("Skipping malformed line %d: %v\n", i+1, record)
			continue
		}
		if powertrain != models.HEV.String() && powertrain != models.PHEV.String() && powertrain != models.BEV.String() && powertrain != models.ICE.String() {
			fmt.Printf("Invalid powertrain: %d: %s\n", i+1, powertrain)
			continue
		}
		if len(strings.Split(definitionID, "_")) != 3 {
			fmt.Printf("Invalid definitionID: %s\n", definitionID)
			notFoundDefinitions = append(notFoundDefinitions, definitionID)
			continue
		}
		fmt.Printf("DefinitionID: %s, Powertrain: %s\n", definitionID, powertrain)

		deviceDefinition, manufID, err := onChainSvc.GetDefinitionByID(ctx, definitionID)
		if err != nil {
			fmt.Printf("%s: Error getting device definition: %v\n", definitionID, err)
			notFoundDefinitions = append(notFoundDefinitions, definitionID)
			continue
		}

		manufName, err := onChainSvc.GetManufacturerNameByID(ctx, manufID)
		if err != nil {
			fmt.Printf("%s: Error getting manufacturer name: %v\n", manufID, err)
			continue
		}
		set := false
		if deviceDefinition.Metadata != nil {
			deviceDefinition.Metadata = &models.DeviceDefinitionMetadata{
				DeviceAttributes: make([]models.DeviceTypeAttribute, 0),
			}
		}
		for i2, attribute := range deviceDefinition.Metadata.DeviceAttributes {
			if attribute.Name == common.PowerTrainType {
				deviceDefinition.Metadata.DeviceAttributes[i2].Value = powertrain
				set = true
				break
			}
		}

		if !set {
			deviceDefinition.Metadata.DeviceAttributes = append(deviceDefinition.Metadata.DeviceAttributes, models.DeviceTypeAttribute{
				Name:  common.PowerTrainType,
				Value: powertrain,
			})
		}
		md, _ := json.Marshal(deviceDefinition.Metadata)
		updateContract := contracts.DeviceDefinitionUpdateInput{
			Id:         deviceDefinition.ID,
			Metadata:   string(md),
			Ksuid:      deviceDefinition.KSUID,
			DeviceType: deviceDefinition.DeviceType,
			ImageURI:   deviceDefinition.ImageURI,
		}

		update, err := onChainSvc.Update(ctx, manufName, updateContract)
		if err != nil {
			fmt.Printf("%s: Error updating device definition: %v\n", definitionID, err)
			if strings.Contains(err.Error(), "nonce too low:") {
				time.Sleep(10 * time.Second)
				update, err = onChainSvc.Update(ctx, manufName, updateContract)
				if err != nil {
					fmt.Printf("%s: Error updating device definition: %v\n", definitionID, err)
				}
			}
			return subcommands.ExitFailure
		}
		fmt.Printf("%s: Updated device definition trx id: %s\nWaiting 10 seconds before next update\n", definitionID, *update)

		time.Sleep(10 * time.Second)
	}

	fmt.Printf("Not found definitions: %d\n", len(notFoundDefinitions))
	for _, definitionID := range notFoundDefinitions {
		fmt.Printf("%s\n", definitionID)
	}

	return subcommands.ExitSuccess
}
