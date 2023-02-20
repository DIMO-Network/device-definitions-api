package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// BulkDecodeVIN godoc
// @Summary gets a list of decoded vins.
// @ID BulkDecodeVIN
// @Description decodes a list of VINs
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /bulk-decode/ [post]
func BulkDecodeVIN(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		payload := make([]string, 0)

		c.BodyParser(&payload)

		command := &commands.BulkValidateVinCommand{VINs: payload}

		result, _ := m.Send(c.UserContext(), command)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}