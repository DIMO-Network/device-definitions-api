package main

import (
	"context"

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
		mediator.WithHandler(&commands.SyncSmartCartCompatibilityCommand{}, commands.NewSyncSmartCartCompatibilityCommandHandler(pdb.DBS, smartCartService, deviceDefinitionRepository)),
	)

	_, _ = m.Send(ctx, &commands.SyncSmartCartForwardCompatibilityCommand{})

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

	_, _ = m.Send(ctx, &commands.SyncSmartCartForwardCompatibilityCommand{})

}

func nhtsaSyncRecalls(ctx context.Context, s *config.Settings, logger zerolog.Logger) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &s.DB, true)
	pdb.WaitForDB(logger)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, s)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, s)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(metricsSvc, &logger, s)),
		mediator.WithHandler(&commands.SyncNHTSARecallsCommand{}, commands.NewSyncNHTSARecallsCommandHandler(pdb.DBS, &s.NHTSARecallsFileURL)),
	)

	_, _ = m.Send(ctx, &commands.SyncNHTSARecallsCommand{})

}
