package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

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

	for i, record := range records {
		if len(record) < 2 {
			fmt.Printf("Skipping malformed line %d: %v\n", i+1, record)
			continue
		}
		definitionID := record[0]
		powertrain := record[1]
		if definitionID == "" || powertrain == "" {
			fmt.Printf("Skipping malformed line %d: %v\n", i+1, record)
			continue
		}
		fmt.Printf("DefinitionID: %s, Powertrain: %s\n", definitionID, powertrain)

		deviceDefinition, manufID, err := onChainSvc.GetDefinitionByID(ctx, definitionID)
		if err != nil {
			fmt.Printf("%s: Error getting device definition: %v\n", definitionID, err)
			continue
		}

		manufName, err := onChainSvc.GetManufacturerNameByID(ctx, manufID)
		if err != nil {
			fmt.Printf("%s: Error getting manufacturer name: %v\n", manufID, err)
			continue
		}
		set := false
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

		update, err := onChainSvc.Update(ctx, manufName, contracts.DeviceDefinitionUpdateInput{
			Id:         deviceDefinition.ID,
			Metadata:   string(md),
			Ksuid:      deviceDefinition.KSUID,
			DeviceType: deviceDefinition.DeviceType,
			ImageURI:   deviceDefinition.ImageURI,
		})
		if err != nil {
			fmt.Printf("%s: Error updating device definition: %v\n", definitionID, err)
			return subcommands.ExitFailure
		}
		fmt.Printf("%s: Updated device definition trx id: %s\n", definitionID, *update)
	}

	return subcommands.ExitSuccess
}
