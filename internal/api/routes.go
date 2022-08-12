package api

import (
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/api/handlers"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

func RegisterUserDeviceRoutes(app fiber.Router, m mediator.Mediator) {
	app.Get("/device-definitions/:id", handlers.GetDeviceDefinitionById(m))
}
