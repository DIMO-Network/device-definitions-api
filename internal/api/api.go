package api

import (
	"context"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"

	"os"
	"os/signal"
	"syscall"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func Run(ctx context.Context, logger zerolog.Logger, settings *config.Settings, send sender.Sender) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &settings.DB, true)
	pdb.WaitForDB(logger)

	// redis
	redisCache := redis.NewRedisCacheService(settings.IsProd(), settings.Redis)

	ethClient, err := ethclient.Dial(settings.EthereumRPCURL.String())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}

	//infra
	drivlyAPIService := gateways.NewDrivlyAPIService(settings)
	vincarioAPIService := gateways.NewVincarioAPIService(settings, &logger)
	fuelAPIService := gateways.NewFuelAPIService(settings, &logger)
	autoIsoAPIService := gateways.NewAutoIsoAPIService(settings)
	ddOnChainService := gateways.NewDeviceDefinitionOnChainService(settings, &logger, ethClient, chainID, send)
	datGroupWSService := gateways.NewDATGroupAPIService(settings, &logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS, ddOnChainService)
	makeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	deviceIntegrationRepository := repositories.NewDeviceIntegrationRepository(pdb.DBS)
	deviceStyleRepository := repositories.NewDeviceStyleRepository(pdb.DBS)
	deviceMakeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	vinRepository := repositories.NewVINRepository(pdb.DBS)

	//cache services
	ddCacheService := services.NewDeviceDefinitionCacheService(redisCache, deviceDefinitionRepository, ddOnChainService)
	vincDecodingService := services.NewVINDecodingService(drivlyAPIService, vincarioAPIService, autoIsoAPIService, &logger, deviceDefinitionRepository, datGroupWSService)
	powerTrainTypeService, err := services.NewPowerTrainTypeService(pdb.DBS, "powertrain_type_rule.yaml", &logger, ddOnChainService)
	searchService := search.NewTypesenseAPIService(settings, &logger)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	//custom commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, settings)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, settings)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, settings)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionQuery{}, queries.NewGetAllDeviceDefinitionGroupQueryHandler(deviceDefinitionRepository, makeRepository)),
		mediator.WithHandler(&queries.GetDevicesMMYQuery{}, queries.NewGetDevicesMMYQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQuery{}, queries.NewGetDeviceDefinitionByIDQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDsQuery{}, queries.NewGetDeviceDefinitionByIDsQueryHandler(ddCacheService, &logger)),
		mediator.WithHandler(&queries.GetDeviceDefinitionWithRelsQuery{}, queries.NewGetDeviceDefinitionWithRelsQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByMakeModelYearQuery{}, queries.NewGetDeviceDefinitionByMakeModelYearQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionBySlugQuery{}, queries.NewGetDeviceDefinitionBySlugQueryHandler(ddOnChainService, pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceDefinitionBySlugNameQuery{}, queries.NewGetDeviceDefinitionBySlugNameQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionBySourceQuery{}, queries.NewGetDeviceDefinitionBySourceQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&queries.GetDeviceDefinitionWithoutImageQuery{}, queries.NewGetDeviceDefinitionWithoutImageQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionQuery{}, queries.NewGetAllDeviceDefinitionQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByDynamicFilterQuery{}, queries.NewGetDeviceDefinitionByDynamicFilterQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetAllIntegrationQuery{}, queries.NewGetAllIntegrationQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationByIDQuery{}, queries.NewGetIntegrationByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByIDQuery{}, queries.NewGetDeviceStyleByIDQueryHandler(pdb.DBS, ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceStyleByFilterQuery{}, queries.NewGetDeviceStyleByFilterQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByDeviceDefinitionIDQuery{}, queries.NewGetDeviceStyleByDeviceDefinitionIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByExternalIDQuery{}, queries.NewGetDeviceStyleByExternalIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceMakeByNameQuery{}, queries.NewGetDeviceMakeByNameQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceMakeBySlugQuery{}, queries.NewGetDeviceMakeBySlugQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceMakeByTokenIDQuery{}, queries.NewGetDeviceMakeByTokenIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetAllDeviceMakeQuery{}, queries.NewGetAllDeviceMakeQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceTypeByIDQuery{}, queries.NewGetDeviceTypeByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceDefinitionImagesByIDsQuery{}, queries.NewGetDeviceDefinitionImagesByIDsQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&commands.CreateDeviceDefinitionCommand{}, commands.NewCreateDeviceDefinitionCommandHandler(deviceDefinitionRepository, pdb.DBS, powerTrainTypeService)),
		mediator.WithHandler(&commands.CreateDeviceIntegrationCommand{}, commands.NewCreateDeviceIntegrationCommandHandler(deviceIntegrationRepository, pdb.DBS, ddCacheService, deviceDefinitionRepository)),
		mediator.WithHandler(&commands.CreateDeviceStyleCommand{}, commands.NewCreateDeviceStyleCommandHandler(deviceStyleRepository, ddCacheService)),
		mediator.WithHandler(&commands.CreateIntegrationCommand{}, commands.NewCreateIntegrationCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceMakeCommand{}, commands.NewCreateDeviceMakeCommandHandler(deviceMakeRepository)),
		mediator.WithHandler(&commands.UpdateDeviceMakeCommand{}, commands.NewUpdateDeviceMakeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateDeviceStyleCommand{}, commands.NewUpdateDeviceStyleCommandHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetAllDeviceTypeQuery{}, queries.NewGetAllDeviceTypeQueryHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateDeviceTypeCommand{}, commands.NewUpdateDeviceTypeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceTypeCommand{}, commands.NewCreateDeviceTypeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.DeleteDeviceTypeCommand{}, commands.NewDeleteDeviceTypeCommandHandler(pdb.DBS)),

		mediator.WithHandler(&queries.GetAllIntegrationFeatureQuery{}, queries.NewGetAllIntegrationFeatureQuery(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationFeatureByIDQuery{}, queries.NewGetIntegrationFeatureByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationOptionsQuery{}, queries.NewGetIntegrationOptionsQueryHandler(pdb.DBS)),

		mediator.WithHandler(&queries.DecodeVINQuery{}, queries.NewDecodeVINQueryHandler(pdb.DBS, vincDecodingService, vinRepository, deviceDefinitionRepository, &logger, fuelAPIService, powerTrainTypeService, ddOnChainService)),

		mediator.WithHandler(&queries.GetAllDeviceDefinitionByMakeYearRangeQuery{}, queries.NewGetAllDeviceDefinitionByMakeYearRangeQueryHandler(deviceDefinitionRepository)),

		mediator.WithHandler(&queries.GetDefinitionsWithHWTemplateQuery{}, queries.NewGetDefinitionsWithHWTemplateQueryHandler(pdb.DBS, &logger)),

		mediator.WithHandler(&commands.BulkValidateVinCommand{}, commands.NewBulkValidateVinCommandHandler(
			pdb.DBS,
			queries.NewDecodeVINQueryHandler(pdb.DBS, vincDecodingService, vinRepository, deviceDefinitionRepository, &logger, fuelAPIService, powerTrainTypeService, ddOnChainService),
			queries.NewGetDeviceDefinitionByIDQueryHandler(ddCacheService),
		)),

		mediator.WithHandler(&queries.GetIntegrationByTokenIDQuery{}, queries.NewGetIntegrationByTokenIDQueryHandler(pdb.DBS, &logger)),

		mediator.WithHandler(&queries.GetAllDeviceDefinitionOnChainQuery{}, queries.NewGetAllDeviceDefinitionOnChainQueryHandler(pdb.DBS, ddOnChainService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionOnChainByIDQuery{}, queries.NewGetDeviceDefinitionOnChainByIDQueryHandler(ddCacheService, pdb.DBS)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionBySearchQuery{}, queries.NewGetAllDeviceDefinitionBySearchQueryHandler(searchService)),
		mediator.WithHandler(&queries.GetR1CompatibilitySearch{}, queries.NewGetR1CompatibilitySearchQueryHandler(searchService)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionByAutocompleteQuery{}, queries.NewGetAllDeviceDefinitionByAutocompleteQueryHandler(searchService)),
		mediator.WithHandler(&queries.GetCompatibilityR1SheetQuery{}, queries.NewCompatibilityR1SheetQueryHandler(settings)),
	)

	//fiber
	app := fiber.New(common.FiberConfig(settings.Environment != "local"))
	app.Use(cors.New())
	app.Use(metrics.HTTPMetricsPrometheusMiddleware)
	app.Use(recover.New())

	// TODO: This line is catching the errors and is not taking the general configuration.
	//app.Use(zflogger.New(logger, nil))

	//routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("device definitions api running!"))
	})
	// Traditional tokens
	jwtAuth := jwtware.New(jwtware.Config{
		JWKSetURLs: []string{settings.JwtKeySetURL},
	})

	RegisterDeviceDefinitionsRoutes(app, *m, jwtAuth)
	RegisterIntegrationRoutes(app, *m)
	RegisterDeviceTypeRoutes(app, *m)
	RegisterDeviceMakesRoutes(app, *m)

	app.Get("/v1/swagger/*", swagger.HandlerDefault)

	go StartGrpcServer(logger, settings, *m, pdb.DBS, ddOnChainService)

	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err).Send()
		}
	}()
	startMonitoringServer(logger, settings)
	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent with length of 1
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel
	<-c                                             // This blocks the main thread until an interrupt is received
	logger.Info().Msg("Gracefully shutting down and running cleanup tasks...")
	_ = ctx.Done()
	_ = app.Shutdown()
	_ = pdb.DBS().Writer.Close()
	_ = pdb.DBS().Reader.Close()
}

// startMonitoringServer start server for monitoring endpoints. Could likely be moved to shared lib.
func startMonitoringServer(logger zerolog.Logger, settings *config.Settings) {
	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		// 8888 is our standard port for exposing metrics in DIMO infra
		if err := monApp.Listen(":" + settings.MonitoringPort); err != nil {
			logger.Fatal().Err(err).Str("port", settings.MonitoringPort).Msg("Failed to start monitoring web server.")
		}
	}()

	logger.Info().Str("port", settings.MonitoringPort).Msg("Started monitoring web server.")
}
