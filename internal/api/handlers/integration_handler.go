package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/gofiber/fiber/v2"
)

// GetDeviceIntegrationsByID godoc
// @Summary gets all the available integrations for a device definition.
// @ID GetDeviceIntegrationsByID
// @Description gets all the available integrations for a device definition. Includes the capabilities of the device with the integration
// @Tags device-definitions
// @Param  id path string true "device definition id"
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /device-definitions/{id}/integrations [get]
func GetDeviceIntegrationsByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := &queries.GetDeviceDefinitionWithRelsQuery{DeviceDefinitionID: id}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetIntegrations godoc
// @Summary gets list of integrations we have defined.
// @ID GetIntegrations
// @Description gets list of integrations we have defined.
// @Tags device-definitions
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /integrations [get]
func GetIntegrations(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		query := &queries.GetAllIntegrationQuery{}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetIntegrationByID godoc
// @Summary gets integration by id.
// @ID GetIntegrationByID
// @Description gets integration by id.
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /integrations/{id} [get]
func GetIntegrationByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		query := &queries.GetIntegrationByIDQuery{IntegrationID: []string{id}}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
