package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elastic"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func searchSyncDD(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)

	//infra
	elasticSearchService, _ := elastic.NewElasticAppSearchService(s, logger)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncSearchDataCommand{}, commands.NewSyncSearchDataCommandHandler(pdb.DBS, elasticSearchService, logger)),
	)

	_, _ = m.Send(ctx, &commands.SyncSearchDataCommand{})
}

func ipfsSyncData(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncIPFSDataCommand{}, commands.NewSyncIPFSDataCommandHandler(pdb.DBS, s.IPFSNodeEndpoint)),
	)

	_, _ = m.Send(ctx, &commands.SyncIPFSDataCommand{})
}

func smartCarCompatibility(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//infra
	smartCartService := gateways.NewSmartCarService(pdb.DBS, logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
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

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
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
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncTeslaIntegrationCommand{}, commands.NewSyncTestlaIntegrationCommandHandler(pdb.DBS, &logger)),
	)

	_, _ = m.Send(ctx, &commands.SyncTeslaIntegrationCommand{})

}

func nhtsaSyncRecalls(ctx context.Context, s *config.Settings, logger zerolog.Logger) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//repos
	deviceNHTSARecallsRepository := repositories.NewDeviceNHTSARecallsRepository(pdb.DBS)
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithHandler(&commands.SyncNHTSARecallsCommand{}, commands.NewSyncNHTSARecallsCommandHandler(pdb.DBS, &logger, deviceNHTSARecallsRepository, deviceDefinitionRepository, &s.NHTSARecallsFileURL)),
	)

	_, _ = m.Send(ctx, &commands.SyncNHTSARecallsCommand{})

}

func vinNumbersSync(ctx context.Context, s *config.Settings, logger zerolog.Logger, args []string) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//infra
	drivlyAPIService := gateways.NewDrivlyAPIService(s)
	vincarioAPIService := gateways.NewVincarioAPIService(s, &logger)

	//service
	vinDecodingService := services.NewVINDecodingService(drivlyAPIService, vincarioAPIService, &logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)
	vinRepository := repositories.NewVINRepository(pdb.DBS)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		mediator.WithHandler(&queries.DecodeVINQuery{}, queries.NewDecodeVINQueryHandler(pdb.DBS, vinDecodingService, vinRepository, deviceDefinitionRepository, &logger)),
	)

	filePath := args[1]
	readFile, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		vin := fileScanner.Text()
		_, _ = m.Send(ctx, &queries.DecodeVINQuery{VIN: vin})
	}

	readFile.Close()

}
