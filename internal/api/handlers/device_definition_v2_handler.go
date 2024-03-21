package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	_ "github.com/DIMO-Network/device-definitions-api/internal/core/models" // required for swagger to generate modesl
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/gofiber/fiber/v2"
)

// GetDeviceDefinitionV2ByID godoc
// @Summary gets a device definition
// @ID GetDeviceDefinitionV2ByID
// @Description gets a device definition
// @Tags device-definitions
// @Param  make path string true "device make name"
// @Param  id path string true "device definition id"
// @Produce json
// @Success 200 {object} models.GetDeviceDefinitionQueryResult
// @Failure 404
// @Failure 500
// @Router /v2/device-definitions/{make}/{id} [get]
func GetDeviceDefinitionV2ByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		make := c.Params("make")
		query := &queries.GetDeviceDefinitionOnChainByIDQuery{DeviceDefinitionID: id, MakeSlug: make}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetDeviceDefinitionV2All godoc
// @Summary gets all device definitions by Makes, models, and years
// @ID GetDeviceDefinitionV2All
// @Description gets a device definition
// @Param  make path string true "device make name"
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 500
// @Router /v2/device-definitions/{make} [get]
func GetDeviceDefinitionV2All(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		make := c.Params("make")
		query := &queries.GetAllDeviceDefinitionOnChainQuery{MakeSlug: make}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
