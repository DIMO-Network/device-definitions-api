package api

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/trace"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/redis"
	"github.com/DIMO-Network/zflogger"
	"github.com/TheFellow/go-mediator/mediator"
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

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)
	makeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	deviceIntegrationRepository := repositories.NewDeviceIntegrationRepository(pdb.DBS)
	deviceStyleRepository := repositories.NewDeviceStyleRepository(pdb.DBS)
	deviceMakeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)

	//cache services
	ddCacheService := services.NewDeviceDefinitionCacheService(redisCache, deviceDefinitionRepository)

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

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&logger, settings)),
		mediator.WithBehaviour(common.NewValidationBehavior(&logger, settings)),
		mediator.WithBehaviour(common.NewErrorHandlingBehavior(&logger, settings)),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionQuery{}, queries.NewGetAllDeviceDefinitionQueryHandler(deviceDefinitionRepository, makeRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQuery{}, queries.NewGetDeviceDefinitionByIDQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIdsQuery{}, queries.NewGetDeviceDefinitionByIdsQueryHandler(ddCacheService, &logger)),
		mediator.WithHandler(&queries.GetDeviceDefinitionWithRelsQuery{}, queries.NewGetDeviceDefinitionWithRelsQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByMakeModelYearQuery{}, queries.NewGetDeviceDefinitionByMakeModelYearQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionBySourceQuery{}, queries.NewGetDeviceDefinitionBySourceQueryHandler(pdb.DBS, &logger)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByDynamicFilterQuery{}, queries.NewGetDeviceDefinitionByDynamicFilterQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetAllIntegrationQuery{}, queries.NewGetAllIntegrationQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetIntegrationByIDQuery{}, queries.NewGetIntegrationByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByIDQuery{}, queries.NewGetDeviceStyleByIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByDeviceDefinitionIDQuery{}, queries.NewGetDeviceStyleByDeviceDefinitionIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&queries.GetDeviceStyleByExternalIDQuery{}, queries.NewGetDeviceStyleByExternalIDQueryHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceDefinitionCommand{}, commands.NewCreateDeviceDefinitionCommandHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&commands.CreateDeviceIntegrationCommand{}, commands.NewCreateDeviceIntegrationCommandHandler(deviceIntegrationRepository)),
		mediator.WithHandler(&commands.CreateDeviceStyleCommand{}, commands.NewCreateDeviceStyleCommandHandler(deviceStyleRepository)),
		mediator.WithHandler(&commands.CreateIntegrationCommand{}, commands.NewCreateIntegrationCommandHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceMakeCommand{}, commands.NewCreateDeviceMakeCommandHandler(deviceMakeRepository)),
		mediator.WithHandler(&commands.UpdateDeviceDefinitionCommand{}, commands.NewUpdateDeviceDefinitionCommandHandler(pdb.DBS, ddCacheService)),
		mediator.WithHandler(&commands.UpdateDeviceDefinitionImageCommand{}, commands.NewUpdateDeviceDefinitionImageCommandHandler(pdb.DBS, ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceCompatibilityQuery{}, queries.NewGetDeviceCompatibilityQueryHandler(pdb.DBS, deviceDefinitionRepository)),
	)

	//fiber
	app := fiber.New(common.FiberConfig(settings.Environment != "local"))

	app.Use(recover.New())
	app.Use(zflogger.New(logger, nil))

	//routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("Welcome dimo api!"))
	})

	RegisterDeviceDefinitionsRoutes(app, *m)
	RegisterIntegrationRoutes(app, *m)

	app.Get("/docs/*", swagger.HandlerDefault)

	go StartGrpcServer(logger, settings, *m)

	// Start Server from a different go routine
	go func() {
		if err := app.Listen(":" + settings.Port); err != nil {
			logger.Fatal().Err(err)
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
