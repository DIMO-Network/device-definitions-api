package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// GetDeviceTypesByID godoc
// @Summary gets a device type.
// @ID GetDeviceTypesByID
// @Description gets a devices type with attributes
// @Tags device-definitions
// @Param  id path string true "device type id"
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /device-types/{id} [get]
func GetDeviceTypesByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := &queries.GetDeviceTypeByIDQuery{DeviceTypeID: id}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
