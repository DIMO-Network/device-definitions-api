package api

import (
	"context"
	"log"

	i_grpc "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/api/grpc"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/config"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/queries"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/trace"
	"github.com/TheFellow/go-mediator/mediator"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Run(ctx context.Context, s *config.Settings) {

	//db
	pdb := db.NewDbConnectionFromSettings(ctx, s, true)

	//infra

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(pdb.DBS)
	makeRepository := repositories.NewDeviceMakeRepository(pdb.DBS)

	//services
	prv, err := trace.NewProvider(trace.ProviderConfig{
		JaegerEndpoint: s.TraceMonitorView,
		ServiceName:    s.ServiceName,
		ServiceVersion: s.ServiceVersion,
		Environment:    s.Environment,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer prv.Close(context.Background())

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.LoggingBehavior{}),
		mediator.WithBehaviour(common.ValidationBehavior{}),
		mediator.WithBehaviour(common.ErrorHandlingBehavior{}),
		mediator.WithHandler(&queries.GetAllDeviceDefinitionQuery{}, queries.NewGetAllDeviceDefinitionQueryHandler(deviceDefinitionRepository, makeRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByIdQuery{}, queries.NewGetDeviceDefinitionByIdQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionWithRelsQuery{}, queries.NewGetDeviceDefinitionWithRelsQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetDeviceDefinitionByMakeModelYearQuery{}, queries.NewGetDeviceDefinitionByMakeModelYearQueryHandler(deviceDefinitionRepository)),
		mediator.WithHandler(&queries.GetAllIntegrationQuery{}, queries.NewGetAllIntegrationQueryHandler(pdb.DBS)),
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

	go i_grpc.StartGrpcServer(s, *m)

	log.Fatal(app.Listen(":" + s.Port))
}
