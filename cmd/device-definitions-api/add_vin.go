package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/pkg/db"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	vinutils "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/subcommands"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type addVINCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	sender sender.Sender
}

func (*addVINCmd) Name() string     { return "addvin" }
func (*addVINCmd) Synopsis() string { return "adds a vin manually to vin_numbers table" }
func (*addVINCmd) Usage() string {
	return `addvin`
}

func (p *addVINCmd) SetFlags(_ *flag.FlagSet) {
}

func (p *addVINCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	vin := ""
	for _, arg := range os.Args {
		if len(arg) == 17 {
			vin = arg
		}
	}
	if len(vin) != 17 {
		fmt.Println("invalid vin, must be 17 chars")
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

	vinDecodeNumber, err := models.FindVinNumber(ctx, pdb.DBS().Reader, vin)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		p.logger.Fatal().Err(err).Send()
		return subcommands.ExitFailure
	}
	if vinDecodeNumber != nil {
		fmt.Println("already registered")
		return subcommands.ExitSuccess
	}
	processedVIN := vinutils.VIN(vin)
	wmi, err := models.Wmis(models.WmiWhere.Wmi.EQ(processedVIN.Wmi())).One(ctx, pdb.DBS().Reader)
	if err != nil {
		fmt.Println("could not find WMI for vin")
		return subcommands.ExitFailure
	}
	vinNumber := models.VinNumber{
		Vin:              vin,
		Wmi:              null.StringFrom(processedVIN.Wmi()),
		VDS:              null.StringFrom(processedVIN.VDS()),
		SerialNumber:     processedVIN.SerialNumber(),
		CheckDigit:       null.StringFrom(processedVIN.CheckDigit()),
		Vis:              null.StringFrom(processedVIN.VIS()),
		ManufacturerName: wmi.ManufacturerName,
		DecodeProvider:   null.StringFrom("manual"),
		Year:             processedVIN.Year(),
	}
	if vinNumber.Year == 0 || vinNumber.Year < 2008 || vinNumber.Year > time.Now().Year() {
		year, err := cmdLineInput("enter model year")
		if err != nil {
			fmt.Println(err.Error())
			return subcommands.ExitFailure
		}
		vinNumber.Year, err = strconv.Atoi(year)
		if err != nil {
			fmt.Println(err.Error())
			return subcommands.ExitFailure
		}
	}

	model, err := cmdLineInput("enter model name as appears in Device Definitions")
	if err != nil {
		fmt.Println(err.Error())
		return subcommands.ExitFailure
	}

	manufacturer, err := onChainSvc.GetManufacturer(stringutils.SlugString(wmi.ManufacturerName))
	if err != nil {
		fmt.Println(err.Error())
		return subcommands.ExitFailure
	}
	definitionID := common.DeviceDefinitionSlug(stringutils.SlugString(wmi.ManufacturerName), stringutils.SlugString(model), int16(vinNumber.Year))
	deviceDefinition, err := onChainSvc.GetDefinitionTableland(ctx, big.NewInt(int64(manufacturer.TokenID)), definitionID)

	if err != nil {
		fmt.Println(err.Error() + " " + model + " " + strconv.Itoa(vinNumber.Year))
		return subcommands.ExitFailure
	}
	vinNumber.DefinitionID = deviceDefinition.ID

	err = vinNumber.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	if err != nil {
		fmt.Println(err.Error())
		return subcommands.ExitFailure
	}

	// improvement, from existing data, try to guess Model
	fmt.Printf("added success, ddId: %s\n", vinNumber.DefinitionID)
	return subcommands.ExitSuccess
}

func cmdLineInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
