package api

import (
	"context"
	"log"

	i_grpc "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/api/grpc"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/features/device_definition/queries"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/trace"
	intShared "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/shared"
	"github.com/TheFellow/go-mediator/mediator"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Run(s intShared.Settings) {

	//db
	sqlDb := db.Connection(s)

	//infra

	//repos
	deviceDefinitionRepository := repositories.NewDeviceDefinitionRepository(sqlDb)

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
		mediator.WithHandler(&queries.GetByIdQuery{}, queries.NewGetByIdQueryHandler(deviceDefinitionRepository)),
	)

	//fiber
	app := fiber.New(common.FiberConfig())

	app.Use(recover.New())

	//routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("Welcome dimo api!"))
	})

	RegisterUserDeviceRoutes(app, *m)
	RegisterIntegrationRoutes(app, *m)

	app.Get("/docs/*", swagger.HandlerDefault)

	go i_grpc.StartGrpcServer(s, *m)

	log.Fatal(app.Listen(":" + s.Port))
}
