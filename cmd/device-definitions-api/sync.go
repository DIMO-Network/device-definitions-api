package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"

	"github.com/ethereum/go-ethereum/ethclient"

	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

type syncOpsCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	vinNumbers bool
}

func (*syncOpsCmd) Name() string { return "sync" }
func (*syncOpsCmd) Synopsis() string {
	return "pick a sync option from the list of supported operations."
}
func (*syncOpsCmd) Usage() string {
	return `sync [-vin-csv]`
}

func (p *syncOpsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.vinNumbers, "vin-csv", false, "decode vins from CSV")
}

func (p *syncOpsCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if p.vinNumbers {
		filename := "tmp/vins.csv"
		if len(f.Args()) > 2 {
			filename = f.Args()[2]
		}
		fmt.Printf("using filename %s to get vins\n", filename)
		vinNumbersSync(ctx, &p.settings, p.logger, filename)
	}

	return subcommands.ExitSuccess
}

// vinNumbersSync reads in the passed in list of vins from the filename and calls third party to decode and insert into our vin_numbers db
func vinNumbersSync(ctx context.Context, s *config.Settings, logger zerolog.Logger, filename string) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	send, err := createSender(ctx, s, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create sender.")
	}

	ethClient, err := ethclient.Dial(s.EthereumRPCURL.String())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}

	queryInstance, err := contracts.NewRegistry(s.EthereumRegistryAddress, ethClient)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry query instance.")
	}

	//infra
	drivlyAPIService := gateways.NewDrivlyAPIService(s)
	vincarioAPIService := gateways.NewVincarioAPIService(s, &logger)
	datGroupWSService := gateways.NewDATGroupAPIService(s, &logger)
	fuelAPIService := gateways.NewFuelAPIService(s, &logger)
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(s, &logger, ethClient, chainID, send)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, deviceDefinitionOnChainService)
	vinRepository := repositories.NewVINRepository(pdb.DBS, queryInstance)

	//service
	vinDecodingService := services.NewVINDecodingService(drivlyAPIService, vincarioAPIService, nil, &logger, deviceDefinitionRepository, datGroupWSService)
	powerTrainTypeService, err := services.NewPowerTrainTypeService(pdb.DBS, "powertrain_type_rule.yaml", &logger, deviceDefinitionOnChainService)
	if err != nil {
		logger.Fatal().Err(err).Stack().Send()
	}

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithHandler(&queries.DecodeVINQuery{}, queries.NewDecodeVINQueryHandler(pdb.DBS, vinDecodingService, vinRepository, deviceDefinitionRepository, &logger, fuelAPIService, powerTrainTypeService, deviceDefinitionOnChainService)),
	)

	readFile, err := os.Open(filename)

	if err != nil {
		fmt.Println(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		vin := fileScanner.Text()
		result, err := m.Send(ctx, &queries.DecodeVINQuery{VIN: vin})
		if err == nil && result != nil {
			r, ok := result.(*p_grpc.DecodeVinResponse)
			if ok {
				fmt.Printf("decoded vin %s, ddID: %s, year: %d, source: %s\n", vin, r.DefinitionId, r.Year, r.Source)
			}
		}
	}

	err = readFile.Close()
	if err != nil {
		fmt.Println(err)
	}
}
