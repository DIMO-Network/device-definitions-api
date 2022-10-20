package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// GetDeviceTypesByID godoc
// @Summary gets a device type.
// @ID GetDeviceIntegrationsByID
// @Description gets a devices type with attributes
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetDeviceTypeQueryResult
// @Failure 404 {object} common.ProblemDetails{}
// @Failure 500 {object} common.ProblemDetails{}
// @Router /device-types/{id} [get]
func GetDeviceTypesByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := &queries.GetDeviceTypeByIDQuery{DeviceTypeID: id}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
