package handlers

import (
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// GetDeviceDefinitionByID godoc
// @Summary gets a device definition
// @ID GetDeviceDefinitionByID
// @Description gets a device definition
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /device-definitions/{id} [get]
func GetDeviceDefinitionByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		query := &queries.GetDeviceDefinitionByIDQuery{DeviceDefinitionID: id}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetDeviceDefinitionAll godoc
// @Summary gets all device definitions by Makes, models, and years
// @ID GetDeviceDefinitionAll
// @Description gets a device definition
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 500
// @Router /device-definitions/all [get]
func GetDeviceDefinitionAll(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		query := &queries.GetAllDeviceDefinitionQuery{}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetDeviceDefinitionByMMY godoc
// @Summary gets a specific device definition by make model and year.
// @ID GetDeviceDefinitionByMMY
// @Description gets a specific device definition by make model and year.
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /device-definitions [get]
func GetDeviceDefinitionByMMY(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		make := c.Query("make")
		model := c.Query("model")
		year := c.Query("year")
		yrInt, _ := strconv.Atoi(year)

		query := &queries.GetDeviceDefinitionByMakeModelYearQuery{Make: make, Model: model, Year: yrInt}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
