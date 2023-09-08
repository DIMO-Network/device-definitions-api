package api

import (
	"context"

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
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elastic"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/trace"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func Run(ctx context.Context, logger zerolog.Logger, settings *config.Settings) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &settings.DB, true)
	pdb.WaitForDB(logger)

	// redis
	redisCache := redis.NewRedisCacheService(settings.Environment == "prod", settings.Redis)

	//infra
	drivlyAPIService := gateways.NewDrivlyAPIService(settings)
	vincarioAPIService := gateways.NewVincarioAPIService(settings, &logger)
	fuelAPIService := gateways.NewFuelAPIService(settings, &logger)
	elasticSearchService, _ := elastic.NewElasticAppSearchService(settings, logger)

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)
	makeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	deviceIntegrationRepository := repositories.NewDeviceIntegrationRepository(pdb.DBS)
	deviceStyleRepository := repositories.NewDeviceStyleRepository(pdb.DBS)
	deviceMakeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	vinRepository := repositories.NewVINRepository(pdb.DBS)

	//cache services
	ddCacheService := services.NewDeviceDefinitionCacheService(redisCache, deviceDefinitionRepository)
	vincDecodingService := services.NewVINDecodingService(drivlyAPIService, vincarioAPIService, &logger, deviceDefinitionRepository)
	powerTrainTypeService, err := services.NewPowerTrainTypeService(pdb.DBS, "powertrain_type_rule.yaml", &logger)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	//services
	prv, err := trace.NewProvider(trace.ProviderConfig{
		JaegerEndpoint: settings.TraceMonitorView,
		ServiceName:    settings.ServiceName,
		ServiceVersion: settings.ServiceVersion,
		Environment:    settings.Environment,
	})
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer prv.Close(context.Background())

	//custom commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&logger, settings)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&logger, settings)),
		//mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, settings)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionQuery{}, queries.NewGetAllDeviceDefinitionGroupQueryHandler(deviceDefinitionRepository, makeRepository)),
		mediator.WithHandler(&queries.GetDevicesMMYQuery{}, queries.NewGetDevicesMMYQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQuery{}, queries.NewGetDeviceDefinitionByIDQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIdsQuery{}, queries.NewGetDeviceDefinitionByIdsQueryHandler(ddCacheService, &logger)),
		mediator.WithHandler(&queries.GetDeviceDefinitionWithRelsQuery{}, queries.NewGetDeviceDefinitionWithRelsQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByMakeModelYearQuery{}, queries.NewGetDeviceDefinitionByMakeModelYearQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionBySlugQuery{}, queries.NewGetDeviceDefinitionBySlugQueryHandler(ddCacheService)),
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
		mediator.WithHandler(&queries.GetRecallsByMakeQuery{}, queries.NewGetRecallsByMakeQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetRecallsByModelQuery{}, queries.NewGetRecallsByModelQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetReviewsDynamicFilterQuery{}, queries.NewGetReviewsDynamicFilterQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetReviewsByDeviceDefinitionIDQuery{}, queries.NewGetReviewsByDeviceDefinitionIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetReviewsByIDQuery{}, queries.NewGetReviewsByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceDefinitionImagesByIdsQuery{}, queries.NewGetDeviceDefinitionImagesByIdsQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&commands.CreateReviewCommand{}, commands.NewCreateReviewCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateReviewCommand{}, commands.NewUpdateReviewCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.DeleteReviewCommand{}, commands.NewDeleteReviewCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.ApproveReviewCommand{}, commands.NewApproveReviewCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceDefinitionCommand{}, commands.NewCreateDeviceDefinitionCommandHandler(deviceDefinitionRepository, pdb.DBS, powerTrainTypeService)),
		mediator.WithHandler(&commands.CreateDeviceIntegrationCommand{}, commands.NewCreateDeviceIntegrationCommandHandler(deviceIntegrationRepository, pdb.DBS, ddCacheService, deviceDefinitionRepository)),
		mediator.WithHandler(&commands.CreateDeviceStyleCommand{}, commands.NewCreateDeviceStyleCommandHandler(deviceStyleRepository, ddCacheService)),
		mediator.WithHandler(&commands.CreateIntegrationCommand{}, commands.NewCreateIntegrationCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceMakeCommand{}, commands.NewCreateDeviceMakeCommandHandler(deviceMakeRepository)),
		mediator.WithHandler(&commands.UpdateDeviceDefinitionCommand{}, commands.NewUpdateDeviceDefinitionCommandHandler(deviceDefinitionRepository, pdb.DBS, ddCacheService)),
		mediator.WithHandler(&commands.UpdateDeviceDefinitionImageCommand{}, commands.NewUpdateDeviceDefinitionImageCommandHandler(pdb.DBS, ddCacheService)),
		mediator.WithHandler(&queries.GetCompatibilitiesByMakeQuery{}, queries.NewGetDeviceCompatibilityQueryHandler(pdb.DBS, deviceDefinitionRepository)),
		mediator.WithHandler(&commands.UpdateDeviceMakeCommand{}, commands.NewUpdateDeviceMakeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateDeviceStyleCommand{}, commands.NewUpdateDeviceStyleCommandHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetAllDeviceTypeQuery{}, queries.NewGetAllDeviceTypeQueryHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateDeviceTypeCommand{}, commands.NewUpdateDeviceTypeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceTypeCommand{}, commands.NewCreateDeviceTypeCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.DeleteDeviceTypeCommand{}, commands.NewDeleteDeviceTypeCommandHandler(pdb.DBS)),

		mediator.WithHandler(&queries.GetAllIntegrationFeatureQuery{}, queries.NewGetAllIntegrationFeatureQuery(pdb.DBS)),
		mediator.WithHandler(&queries.GetCompatibilityByDeviceDefinitionQuery{}, queries.NewGetCompatibilityByDeviceDefinitionQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetCompatibilityByDeviceDefinitionArrayQuery{}, queries.NewGetCompatibilityByDeviceDefinitionArrayQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationFeatureByIDQuery{}, queries.NewGetIntegrationFeatureByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationOptionsQuery{}, queries.NewGetIntegrationOptionsQueryHandler(pdb.DBS)),

		mediator.WithHandler(&commands.CreateIntegrationFeatureCommand{}, commands.NewCreateIntegrationFeatureCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.UpdateIntegrationFeatureCommand{}, commands.NewUpdateIntegrationFeatureCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.DeleteIntegrationFeatureCommand{}, commands.NewDeleteIntegrationFeatureCommandHandler(pdb.DBS)),
		mediator.WithHandler(&queries.DecodeVINQuery{}, queries.NewDecodeVINQueryHandler(pdb.DBS, vincDecodingService, vinRepository, deviceDefinitionRepository, &logger, fuelAPIService, powerTrainTypeService)),

		mediator.WithHandler(&queries.GetDefinitionsWithHWTemplateQuery{}, queries.NewGetDefinitionsWithHWTemplateQueryHandler(pdb.DBS, &logger)),

		mediator.WithHandler(&commands.BulkValidateVinCommand{}, commands.NewBulkValidateVinCommandHandler(
			pdb.DBS,
			queries.NewDecodeVINQueryHandler(pdb.DBS, vincDecodingService, vinRepository, deviceDefinitionRepository, &logger, fuelAPIService, powerTrainTypeService),
			queries.NewGetCompatibilityByDeviceDefinitionQueryHandler(pdb.DBS),
			queries.NewGetDeviceDefinitionByIDQueryHandler(ddCacheService),
		)),

		mediator.WithHandler(&commands.SyncSearchDataCommand{}, commands.NewSyncSearchDataCommandHandler(pdb.DBS, elasticSearchService, logger)),
		mediator.WithHandler(&queries.GetIntegrationByTokenIDQuery{}, queries.NewGetIntegrationByTokenIDQueryHandler(pdb.DBS, &logger)),
	)

	//fiber
	app := fiber.New(common.FiberConfig(settings.Environment != "local"))

	app.Use(metrics.HTTPMetricsPrometheusMiddleware)
	app.Use(recover.New())

	// TODO: This line is catching the errors and is not taking the general configuration.
	//app.Use(zflogger.New(logger, nil))

	//routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("Welcome dimo api!"))
	})

	RegisterDeviceDefinitionsRoutes(app, *m)
	RegisterIntegrationRoutes(app, *m)
	RegisterDeviceTypeRoutes(app, *m)
	RegisterDeviceMakesRoutes(app, *m)
	RegisterVINRoutes(app, *m)

	app.Get("/docs/*", swagger.HandlerDefault)

	go StartGrpcServer(logger, settings, *m)

	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err).Send()
		}
	}()
	startMonitoringServer(logger)
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
func startMonitoringServer(logger zerolog.Logger) {
	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		// 8888 is our standard port for exposing metrics in DIMO infra
		if err := monApp.Listen(":8888"); err != nil {
			logger.Fatal().Err(err).Str("port", "8888").Msg("Failed to start monitoring web server.")
		}
	}()

	logger.Info().Str("port", "8888").Msg("Started monitoring web server.")
}
