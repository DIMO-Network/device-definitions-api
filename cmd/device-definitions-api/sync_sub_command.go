package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/google/subcommands"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elastic"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

type syncOpsCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	searchSync            bool
	ipfs                  bool
	smartCarCompatibility bool
	smartCarSync          bool
	createTesla           bool
	nhtsa                 bool
	vinNumbers            bool
}

func (*syncOpsCmd) Name() string { return "sync" }
func (*syncOpsCmd) Synopsis() string {
	return "pick a sync option from the list of supported operations."
}
func (*syncOpsCmd) Usage() string {
	return `sync [-search-sync-dds|-ipfs-sync-data|-smartcar-compatibility|-create-tesla-integrations|-nhtsa-sync-recalls]`
}

func (p *syncOpsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.searchSync, "search-sync-dds", false, "search sync dds")
	f.BoolVar(&p.ipfs, "ipfs-sync-data", false, "ipfs sync data")
	f.BoolVar(&p.smartCarCompatibility, "smartcar-compatibility", false, "smartcar compatibility")
	f.BoolVar(&p.smartCarSync, "smartcar-sync", false, "smartcar sync")
	f.BoolVar(&p.createTesla, "create-tesla-integrations", false, "create tesla integrations")
	f.BoolVar(&p.nhtsa, "nhtsa-sync-recalls", false, "nhtsa sync recalls")
	f.BoolVar(&p.vinNumbers, "vin-numbers-sync", false, "vin numbers sync data")
}

func (p *syncOpsCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if p.searchSync {
		searchSyncDD(ctx, &p.settings, p.logger)
	}
	if p.smartCarCompatibility {
		smartCarCompatibility(ctx, &p.settings, p.logger)
	}
	if p.smartCarSync {
		smartCarSync(ctx, &p.settings, p.logger)
	}
	if p.createTesla {
		teslaIntegrationSync(ctx, &p.settings, p.logger)
	}
	if p.nhtsa {
		nhtsaSyncRecalls(ctx, &p.settings, p.logger)
	}

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

func searchSyncDD(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)

	//infra
	elasticSearchService, _ := elastic.NewElasticAppSearchService(s, logger)

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncSearchDataCommand{}, commands.NewSyncSearchDataCommandHandler(pdb.DBS, elasticSearchService, logger)),
	)

	_, _ = m.Send(ctx, &commands.SyncSearchDataCommand{})
}

func smartCarCompatibility(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//infra
	smartCartService := gateways.NewSmartCarService(pdb.DBS, logger)
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(s, &logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, deviceDefinitionOnChainService)

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncSmartCartForwardCompatibilityCommand{},
			commands.NewSyncSmartCartForwardCompatibilityCommandHandler(pdb.DBS, smartCartService, deviceDefinitionRepository)),
	)

	_, _ = m.Send(ctx, &commands.SyncSmartCartForwardCompatibilityCommand{})
}

func smartCarSync(ctx context.Context, s *config.Settings, logger zerolog.Logger) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//infra
	smartCartService := gateways.NewSmartCarService(pdb.DBS, logger)
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(s, &logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, deviceDefinitionOnChainService)

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncSmartCartCompatibilityCommand{},
			commands.NewSyncSmartCartCompatibilityCommandHandler(pdb.DBS, smartCartService, deviceDefinitionRepository)),
	)

	_, _ = m.Send(ctx, &commands.SyncSmartCartCompatibilityCommand{})

}

func teslaIntegrationSync(ctx context.Context, s *config.Settings, logger zerolog.Logger) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncTeslaIntegrationCommand{}, commands.NewSyncTestlaIntegrationCommandHandler(pdb.DBS, &logger)),
	)

	_, _ = m.Send(ctx, &commands.SyncTeslaIntegrationCommand{})

}

func nhtsaSyncRecalls(ctx context.Context, s *config.Settings, logger zerolog.Logger) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(s, &logger)

	//repos
	deviceNHTSARecallsRepository := repositories.NewDeviceNHTSARecallsRepository(pdb.DBS)
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, deviceDefinitionOnChainService)

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncNHTSARecallsCommand{}, commands.NewSyncNHTSARecallsCommandHandler(pdb.DBS, &logger, deviceNHTSARecallsRepository, deviceDefinitionRepository, &s.NHTSARecallsFileURL)),
	)

	_, _ = m.Send(ctx, &commands.SyncNHTSARecallsCommand{})

}

// vinNumbersSync reads in the passed in list of vins from the filename and calls third party to decode and insert into our vin_numbers db
func vinNumbersSync(ctx context.Context, s *config.Settings, logger zerolog.Logger, filename string) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//infra
	drivlyAPIService := gateways.NewDrivlyAPIService(s)
	vincarioAPIService := gateways.NewVincarioAPIService(s, &logger)
	datGroupWSService := gateways.NewDATGroupAPIService(s, &logger)
	fuelAPIService := gateways.NewFuelAPIService(s, &logger)
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(s, &logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, deviceDefinitionOnChainService)
	vinRepository := repositories.NewVINRepository(pdb.DBS)

	//service
	vinDecodingService := services.NewVINDecodingService(drivlyAPIService, vincarioAPIService, nil, &logger, deviceDefinitionRepository, datGroupWSService)
	powerTrainTypeService, err := services.NewPowerTrainTypeService(pdb.DBS, "powertrain_type_rule.yaml", &logger)
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
				fmt.Printf("decoded vin %s, ddID: %s, year: %d, source: %s\n", vin, r.DeviceDefinitionId, r.Year, r.Source)
			}
		}
	}

	readFile.Close()

}
