package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"

	_ "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
)

// GetDeviceMakes godoc
// @Summary gets all device makes
// @ID GetDeviceMakes
// @Description gets all device makes
// @Tags device-definitions
// @Produce json
// @Success 200 {object} grpc.GetDeviceMakeResponse
// @Failure 500
// @Router /device-makes [get]
func GetDeviceMakes(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		query := &queries.GetAllDeviceMakeQuery{}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
