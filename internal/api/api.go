package api

import (
	"context"
	"log"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/trace"
	"github.com/TheFellow/go-mediator/mediator"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
)

func Run(ctx context.Context, logger zerolog.Logger, settings *config.Settings) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, settings, true)
	pdb.WaitForDB(logger)

	// redis
	redisCache := gateways.NewRedisCacheService(settings, 1)

	//infra
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)

	//repos
	makeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)
	deviceIntegrationRepository := repositories.NewDeviceIntegrationRepository(pdb.DBS)

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
		mediator.WithBehaviour(common.LoggingBehavior{}),
		mediator.WithBehaviour(common.ValidationBehavior{}),
		mediator.WithBehaviour(common.ErrorHandlingBehavior{}),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionQuery{}, queries.NewGetAllDeviceDefinitionQueryHandler(deviceDefinitionRepository, makeRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQuery{}, queries.NewGetDeviceDefinitionByIDQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIdsQuery{}, queries.NewGetDeviceDefinitionByIdsQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetDeviceDefinitionWithRelsQuery{}, queries.NewGetDeviceDefinitionWithRelsQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByMakeModelYearQuery{}, queries.NewGetDeviceDefinitionByMakeModelYearQueryHandler(ddCacheService)),
		mediator.WithHandler(&queries.GetAllIntegrationQuery{}, queries.NewGetAllIntegrationQueryHandler(pdb.DBS)),
		mediator.WithHandler(&commands.CreateDeviceDefinitionCommand{}, commands.NewCreateDeviceDefinitionCommandHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&commands.CreateDeviceIntegrationCommand{}, commands.NewCreateDeviceIntegrationCommandHandler(deviceIntegrationRepository)),
		mediator.WithHandler(&commands.UpdateDeviceDefinitionCommand{}, commands.NewUpdateDeviceDefinitionCommandHandler(pdb.DBS, ddCacheService)),
	)

	//fiber
	app := fiber.New(common.FiberConfig())

	app.Use(recover.New())

	//routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("Welcome dimo api!"))
	})

	RegisterDeviceDefinitionsRoutes(app, *m)
	RegisterIntegrationRoutes(app, *m)

	app.Get("/docs/*", swagger.HandlerDefault)

	go StartGrpcServer(settings, *m)

	log.Fatal(app.Listen(":" + settings.Port))
}
