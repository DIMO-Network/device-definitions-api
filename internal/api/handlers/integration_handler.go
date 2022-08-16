package handlers

import (
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/features/integrations/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// @Summary gets all the available integrations for a device definition.
// @ID GetDeviceIntegrationsByID
// @Description gets all the available integrations for a device definition. Includes the capabilities of the device with the integration
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetByDeviceDefinitionIntegrationIdQueryResult
// @Failure 404 {object} common.ProblemDetails{}
// @Failure 500 {object} common.ProblemDetails{}
// @Router /device-definitions/{id}/integrations [get]
func GetDeviceIntegrationsByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := &queries.GetByDeviceDefinitionIntegationIdQuery{DeviceDefinitionID: id}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// @Summary gets list of integrations we have defined.
// @ID GetDeviceIntegrationsByID
// @Description gets list of integrations we have defined.
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetAllQueryResult
// @Failure 404 {object} common.ProblemDetails{}
// @Failure 500 {object} common.ProblemDetails{}
// @Router /integrations [get]
func GetIntegrations(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		query := &queries.GetAllQuery{}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
