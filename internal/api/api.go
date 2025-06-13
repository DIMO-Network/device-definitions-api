package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
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
	"github.com/DIMO-Network/shared/pkg/db"
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
	//redisCache := redis.NewRedisCacheService(settings.IsProd(), settings.Redis)

	ethClient, err := ethclient.Dial(settings.EthereumRPCURL.String())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}

	//infra
	identityAPI := gateways.NewIdentityAPIService(&logger, settings)
	drivlyAPIService := gateways.NewDrivlyAPIService(settings)
	vincarioAPIService := gateways.NewVincarioAPIService(settings, &logger)
	fuelAPIService := gateways.NewFuelAPIService(settings, &logger)
	autoIsoAPIService := gateways.NewAutoIsoAPIService(settings.AutoIsoAPIUid, settings.AutoIsoAPIKey)
	japan17VINAPI := gateways.NewJapan17VINAPI(&logger, settings)
	registryInstance, err := contracts.NewRegistry(settings.EthereumRegistryAddress, ethClient)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create registry query instance.")
	}
	ddOnChainService := gateways.NewDeviceDefinitionOnChainService(settings, &logger, ethClient, chainID, send, pdb.DBS)
	datGroupWSService := gateways.NewDATGroupAPIService(settings, &logger)

	//repos
	//makeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	deviceStyleRepository := repositories.NewDeviceStyleRepository(pdb.DBS)
	vinRepository := repositories.NewVINRepository(pdb.DBS, registryInstance, identityAPI)

	//cache services
	vincDecodingService := services.NewVINDecodingService(drivlyAPIService, vincarioAPIService, autoIsoAPIService, &logger, ddOnChainService, datGroupWSService, pdb.DBS, japan17VINAPI)
	powerTrainTypeService, err := services.NewPowerTrainTypeService("powertrain_type_rule.yaml", &logger, ddOnChainService)
	searchService := search.NewTypesenseAPIService(settings, &logger)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	decodeVINHandler := queries.NewDecodeVINQueryHandler(pdb.DBS, vincDecodingService, vinRepository, &logger, fuelAPIService, powerTrainTypeService, ddOnChainService, identityAPI)
	upsertVINHandler := queries.NewUpsertDecodingQueryHandler(pdb.DBS, &logger, ddOnChainService)

	//custom commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, settings)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, settings)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, settings)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQuery{}, queries.NewGetDeviceDefinitionByIDQueryHandler(ddOnChainService, pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByMakeModelYearQuery{}, queries.NewGetDeviceDefinitionByMakeModelYearQueryHandler(ddOnChainService, pdb.DBS, identityAPI)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByDynamicFilterQuery{}, queries.NewGetDeviceDefinitionByDynamicFilterQueryHandler(pdb.DBS, ddOnChainService)),
		mediator.WithHandler(&queries.GetAllIntegrationQuery{}, queries.NewGetAllIntegrationQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationByIDQuery{}, queries.NewGetIntegrationByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByIDQuery{}, queries.NewGetDeviceStyleByIDQueryHandler(pdb.DBS, ddOnChainService)),
		mediator.WithHandler(&queries.GetDeviceStyleByFilterQuery{}, queries.NewGetDeviceStyleByFilterQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByDeviceDefinitionIDQuery{}, queries.NewGetDeviceStyleByDeviceDefinitionIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByExternalIDQuery{}, queries.NewGetDeviceStyleByExternalIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceTypeByIDQuery{}, queries.NewGetDeviceTypeByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceDefinitionImagesByIDsQuery{}, queries.NewGetDeviceDefinitionImagesByIDsQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&commands.CreateDeviceDefinitionCommand{}, commands.NewCreateDeviceDefinitionCommandHandler(ddOnChainService, pdb.DBS, powerTrainTypeService, fuelAPIService, &logger, identityAPI)),
		mediator.WithHandler(&commands.CreateDeviceStyleCommand{}, commands.NewCreateDeviceStyleCommandHandler(deviceStyleRepository)),
		mediator.WithHandler(&commands.UpdateDeviceStyleCommand{}, commands.NewUpdateDeviceStyleCommandHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetAllDeviceTypeQuery{}, queries.NewGetAllDeviceTypeQueryHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateDeviceTypeCommand{}, commands.NewUpdateDeviceTypeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceTypeCommand{}, commands.NewCreateDeviceTypeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.DeleteDeviceTypeCommand{}, commands.NewDeleteDeviceTypeCommandHandler(pdb.DBS)),

		mediator.WithHandler(&queries.GetIntegrationOptionsQuery{}, queries.NewGetIntegrationOptionsQueryHandler(pdb.DBS)),
		mediator.WithHandler(&commands.BulkValidateVinCommand{}, commands.NewBulkValidateVinCommandHandler(
			pdb.DBS,
			queries.NewDecodeVINQueryHandler(pdb.DBS, vincDecodingService, vinRepository, &logger, fuelAPIService, powerTrainTypeService, ddOnChainService, identityAPI),
			queries.NewGetDeviceDefinitionByIDQueryHandler(ddOnChainService, pdb.DBS),
		)),

		mediator.WithHandler(&queries.GetIntegrationByTokenIDQuery{}, queries.NewGetIntegrationByTokenIDQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionBySearchQuery{}, queries.NewGetAllDeviceDefinitionBySearchQueryHandler(searchService)),
		mediator.WithHandler(&queries.GetR1CompatibilitySearch{}, queries.NewGetR1CompatibilitySearchQueryHandler(searchService)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionByAutocompleteQuery{}, queries.NewGetAllDeviceDefinitionByAutocompleteQueryHandler(searchService)),
		mediator.WithHandler(&queries.GetCompatibilityR1SheetQuery{}, queries.NewCompatibilityR1SheetQueryHandler(settings)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQueryV2{}, queries.NewGetDeviceDefinitionByIDQueryV2Handler(ddOnChainService, pdb.DBS)),
		mediator.WithHandler(&queries.GetVINProfileQuery{}, queries.NewGetVINProfileQueryHandler(pdb.DBS, &logger)),

		mediator.WithHandler(&queries.DecodeVINQuery{}, decodeVINHandler),
		mediator.WithHandler(&queries.UpsertDecodingQuery{}, upsertVINHandler),
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

	app.Get("/v1/swagger/*", swagger.HandlerDefault)

	go StartGrpcServer(logger, settings, *m, pdb.DBS, ddOnChainService, registryInstance, identityAPI, &decodeVINHandler, &upsertVINHandler)

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
