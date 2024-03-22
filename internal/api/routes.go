package api

import (
	"github.com/DIMO-Network/device-definitions-api/internal/api/handlers"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/gofiber/fiber/v2"
)

func RegisterDeviceDefinitionsRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-definitions/all", handlers.GetDeviceDefinitionAll(m)).Name("device-definitions-all")
	app.Get("/device-definitions/:id", handlers.GetDeviceDefinitionByID(m)).Name("device-definitions-by-id")
	app.Get("/device-definitions/:id/integrations", handlers.GetDeviceIntegrationsByID(m)).Name("device-definitions-with-integrations")
	app.Get("/device-definitions", handlers.GetDeviceDefinitionByMMY(m)).Name("device-definitions")

	app.Get("/v2/device-definitions/:make/all", handlers.GetDeviceDefinitionV2All(m)).Name("device-definitions-all-v2")
	app.Get("/v2/device-definitions/:make/:id", handlers.GetDeviceDefinitionV2ByID(m)).Name("device-definitions-by-id-v2")

}

func RegisterIntegrationRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/integrations", handlers.GetIntegrations(m)).Name("integrations")
	app.Get("/integrations/:id", handlers.GetIntegrationByID(m)).Name("integrations-by-id")
}

func RegisterDeviceTypeRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-types", handlers.GetDeviceTypes(m)).Name("device-types-all")
	app.Get("/device-types/:id", handlers.GetDeviceTypesByID(m)).Name("device-types")
}

func RegisterDeviceMakesRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-makes", handlers.GetDeviceMakes(m)).Name("device-makes")
}

func RegisterVINRoutes(app fiber.Router, m mediator.Mediator) {
	app.Post("/bulk-decode", handlers.BulkDecodeVIN(m)).Name("bulk-decode")
	app.Post("/bulk-decode/csv", handlers.BulkDecodeVINCsv(m)).Name("bulk-decode-csv")
}
