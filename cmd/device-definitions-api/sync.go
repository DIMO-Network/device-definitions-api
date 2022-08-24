package main

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/TheFellow/go-mediator/mediator"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func search_sync_dds(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, s, true)

	//infra
	elasticSearchService, _ := gateways.NewElasticSearchService(s, logger)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.LoggingBehavior{}),
		mediator.WithBehaviour(common.ValidationBehavior{}),
		mediator.WithBehaviour(common.ErrorHandlingBehavior{}),
		mediator.WithHandler(&commands.SyncSearchDataCommand{}, commands.NewSyncSearchDataCommandHandler(pdb.DBS, elasticSearchService)),
	)

	m.Send(ctx, &commands.SyncSearchDataCommand{})
}

func ipfs_sync_data(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, s, true)
	pdb.WaitForDB(logger)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.LoggingBehavior{}),
		mediator.WithBehaviour(common.ValidationBehavior{}),
		mediator.WithBehaviour(common.ErrorHandlingBehavior{}),
		mediator.WithHandler(&commands.SyncIPFSDataCommand{}, commands.NewSyncIPFSDataCommandHandler(pdb.DBS, s.IPFSNodeEndpoint)),
	)

	m.Send(ctx, &commands.SyncIPFSDataCommand{})
}

func smartcar_compatibility(ctx context.Context, s *config.Settings, logger zerolog.Logger) {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, s, true)
	pdb.WaitForDB(logger)

	//infra
	smartCartService := gateways.NewSmartCarService(pdb.DBS, logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.LoggingBehavior{}),
		mediator.WithBehaviour(common.ValidationBehavior{}),
		mediator.WithBehaviour(common.ErrorHandlingBehavior{}),
		mediator.WithHandler(&commands.SyncSmartCartForwardCompatibilityCommand{},
			commands.NewSyncSmartCartForwardCompatibilityCommandHandler(pdb.DBS, smartCartService, deviceDefinitionRepository)),
	)

	m.Send(ctx, &commands.SyncSmartCartForwardCompatibilityCommand{})
}

func smartcar_sync(ctx context.Context, s *config.Settings, logger zerolog.Logger) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, s, true)
	pdb.WaitForDB(logger)

	//infra
	smartCartService := gateways.NewSmartCarService(pdb.DBS, logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.LoggingBehavior{}),
		mediator.WithBehaviour(common.ValidationBehavior{}),
		mediator.WithBehaviour(common.ErrorHandlingBehavior{}),
		mediator.WithHandler(&commands.SyncSmartCartCompatibilityCommand{}, commands.NewSyncSmartCartCompatibilityCommandHandler(pdb.DBS, smartCartService, deviceDefinitionRepository)),
	)

	m.Send(ctx, &commands.SyncSmartCartForwardCompatibilityCommand{})

}
