package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/shared/pkg/logfields"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	vinutils "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/subcommands"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type addVINsCSVCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	sender   sender.Sender
	identity gateways.IdentityAPI
}

func (*addVINsCSVCmd) Name() string { return "addvinscsv" }
func (*addVINsCSVCmd) Synopsis() string {
	return "bulk adds VINs from CSV text with VIN and DefinitionId columns"
}
func (*addVINsCSVCmd) Usage() string {
	return `addvinscsv:
  Reads CSV text from stdin with columns: VIN, DefinitionId
  Example:
    cat vins.csv | go run . addvinscsv
  Or:
    echo "VIN,DefinitionId
    1HGBH41JXMN109186,some-definition-id
    2HGFC2F59FH123456,another-definition-id" | go run . addvinscsv
`
}

func (p *addVINsCSVCmd) SetFlags(_ *flag.FlagSet) {
}

func (p *addVINsCSVCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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

	// Read CSV from stdin
	reader := csv.NewReader(os.Stdin)
	reader.TrimLeadingSpace = true

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to read CSV headers")
		fmt.Println("Failed to read CSV headers:", err)
		return subcommands.ExitFailure
	}

	// Find column indices
	vinIdx := -1
	defIDIdx := -1
	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		switch headerLower {
		case "vin":
			vinIdx = i
		case "definitionid":
			defIDIdx = i
		}
	}

	if vinIdx == -1 || defIDIdx == -1 {
		p.logger.Error().Msg("CSV must contain 'VIN' and 'DefinitionId' columns")
		fmt.Println("Error: CSV must contain 'VIN' and 'DefinitionId' columns")
		return subcommands.ExitFailure
	}

	successCount := 0
	skipCount := 0
	errorCount := 0

	// Process each row
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to read CSV row")
			fmt.Println("Failed to read CSV row:", err)
			errorCount++
			continue
		}

		if len(record) <= vinIdx || len(record) <= defIDIdx {
			p.logger.Error().Msg("Invalid CSV row: missing columns")
			fmt.Println("Invalid CSV row: missing columns")
			errorCount++
			continue
		}

		vin := strings.TrimSpace(record[vinIdx])
		definitionID := strings.TrimSpace(record[defIDIdx])

		if len(vin) != 17 {
			p.logger.Error().Str("vin", vin).Msg("Invalid VIN: must be 17 characters")
			fmt.Printf("Skipping invalid VIN '%s': must be 17 characters\n", vin)
			errorCount++
			continue
		}

		// Check if VIN already exists
		vinDecodeNumber, err := models.FindVinNumber(ctx, pdb.DBS().Reader, vin)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			p.logger.Error().Err(err).Str("vin", vin).Msg("Database error checking VIN")
			fmt.Printf("Error checking VIN '%s': %v\n", vin, err)
			errorCount++
			continue
		}
		if vinDecodeNumber != nil {
			p.logger.Info().Str("vin", vin).Msg("VIN already registered, skipping")
			fmt.Printf("VIN '%s' already registered, skipping\n", vin)
			skipCount++
			continue
		}

		// Process VIN
		processedVIN := vinutils.VIN(vin)
		wmi, err := models.Wmis(models.WmiWhere.Wmi.EQ(processedVIN.Wmi())).One(ctx, pdb.DBS().Reader)
		if err != nil {
			p.logger.Error().Err(err).Str("vin", vin).Str("wmi", processedVIN.Wmi()).Msg("Could not find WMI for VIN")
			fmt.Printf("Error: Could not find WMI '%s' for VIN '%s'\n", processedVIN.Wmi(), vin)
			errorCount++
			continue
		}

		// Verify the device definition exists
		deviceDefinition, err := p.identity.GetDeviceDefinitionByID(definitionID)
		if err != nil || deviceDefinition == nil {
			if err != nil {
				p.logger.Error().Err(err).Str(logfields.DefinitionID, definitionID).Msg("Could not find definition")
			}
			fmt.Printf("Error: Could not find manufacturer '%s' for VIN '%s': %v\n", wmi.ManufacturerName, vin, err)
			errorCount++
			continue
		}

		vinNumber := models.VinNumber{
			Vin:              vin,
			Wmi:              null.StringFrom(processedVIN.Wmi()),
			VDS:              null.StringFrom(processedVIN.VDS()),
			SerialNumber:     processedVIN.SerialNumber(),
			CheckDigit:       null.StringFrom(processedVIN.CheckDigit()),
			Vis:              null.StringFrom(processedVIN.VIS()),
			ManufacturerName: wmi.ManufacturerName,
			DecodeProvider:   null.StringFrom("csv-import"),
			DefinitionID:     definitionID,
			Year:             deviceDefinition.Year,
		}

		// Insert into database
		err = vinNumber.Insert(ctx, pdb.DBS().Writer, boil.Infer())
		if err != nil {
			p.logger.Error().Err(err).Str("vin", vin).Msg("Failed to insert VIN")
			fmt.Printf("Error inserting VIN '%s': %v\n", vin, err)
			errorCount++
			continue
		}

		p.logger.Info().Str("vin", vin).Str("definitionID", vinNumber.DefinitionID).Msg("Successfully added VIN")
		fmt.Printf("✓ Added VIN '%s' with definition ID '%s'\n", vin, vinNumber.DefinitionID)
		successCount++
	}

	// Print summary
	fmt.Println("\n=== Summary ===")
	fmt.Printf("Successfully added: %d\n", successCount)
	fmt.Printf("Skipped (already exist): %d\n", skipCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Printf("Total processed: %d\n", successCount+skipCount+errorCount)

	if errorCount > 0 {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
