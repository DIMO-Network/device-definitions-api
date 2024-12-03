package api

import (
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/api/handlers"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
)

func RegisterDeviceDefinitionsRoutes(app fiber.Router, m mediator.Mediator, auth fiber.Handler) {
	app.Get("/device-definitions/search", handlers.GetDeviceDefinitionSearch(m)).Name("device-definitions-search")
	app.Get("/device-definitions/search-r1", handlers.GetR1CompatibilitySearch(m)).Name("r1-compatibility-search")
	app.Get("/compatibility/r1-sheet", cache.New(cache.Config{Expiration: 5 * time.Minute}),
		handlers.GetCompatibilityR1Sheet(m)).Name("compatibility-r1-sheet")
	app.Get("/device-definitions/autocomplete", handlers.GetDeviceDefinitionAutocomplete(m)).Name("device-definitions-autocomplete")
	app.Get("/device-definitions/all", handlers.GetDeviceDefinitionAll(m)).Name("device-definitions-all")
	app.Get("/device-definitions/:id", handlers.GetDeviceDefinitionByID(m)).Name("device-definitions-by-id")
	// todo deprecate: last 7 days not seeing any calls to this
	app.Get("/device-definitions/:id/integrations", handlers.GetDeviceIntegrationsByID(m)).Name("device-definitions-with-integrations")
	app.Get("/device-definitions", handlers.GetDeviceDefinitionByMMY(m)).Name("device-definitions")

	app.Get("/v2/device-definitions/:make/all", handlers.GetDeviceDefinitionV2All(m)).Name("device-definitions-all-v2")
	app.Get("/v2/device-definitions/:make/:id", handlers.GetDeviceDefinitionV2ByID(m)).Name("device-definitions-by-id-v2")
	// oems by external integration, used by mobile app
	app.Get("/manufacturers/integrations/smartcar", handlers.GetSmartcarManufacturers()).Name("device-definitions-smartcar")

	app.Post("/device-definitions/decode-vin", auth, handlers.DecodeVIN(m)).Name("device-definitions-decode-vin")

	app.Get("/v2/device-definitions/:id", handlers.GetDeviceDefinitionByIDv2(m)).Name("device-definitions-by-idv2")
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
