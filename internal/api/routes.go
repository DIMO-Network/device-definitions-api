package api

import (
	"github.com/DIMO-Network/device-definitions-api/internal/api/handlers"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

func RegisterDeviceDefinitionsRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-definitions/all", handlers.GetDeviceDefinitionAll(m))
	app.Get("/device-definitions/:id", handlers.GetDeviceDefinitionByID(m))
	app.Get("/device-definitions/:id/integrations", handlers.GetDeviceIntegrationsByID(m))
	app.Get("/device-definitions", handlers.GetDeviceDefinitionByMMY(m))
}

func RegisterIntegrationRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/integrations", handlers.GetIntegrations(m))
	app.Get("/integrations/:id", handlers.GetIntegrationByID(m))
}

func RegisterDeviceTypeRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-types/:id", handlers.GetDeviceTypesByID(m))
}

func RegisterDeviceMakesRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-makes", handlers.GetDeviceMakes(m))
}
