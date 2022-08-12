package handlers

import (
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/features/device_definition/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// @Summary gets a device definition
// @ID GetById
// @Description gets a device definition
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetByIdQueryResult
// @Failure 404 {object} common.ProblemDetails{}
// @Failure 500 {object} common.ProblemDetails{}
// @Router /device-definitions/{id} [get]
func GetDeviceDefinitionById(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := &queries.GetByIdQuery{DeviceDefinitionID: id}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
