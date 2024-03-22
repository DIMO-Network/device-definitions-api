package handlers

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	_ "github.com/DIMO-Network/device-definitions-api/internal/core/models" // required for swagger to generate modesl
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/gofiber/fiber/v2"
	"strconv"
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
// @Router /v2/device-definitions/{make}/all [get]
func GetDeviceDefinitionV2All(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		make := c.Params("make")
		pageIndexStr := c.Params("pageIndex")
		pageSizeStr := c.Params("pageSize")

		var pageIndex int32 = 1
		if pageIndexStr != "" {
			pageIndex64, err := strconv.ParseInt(pageIndexStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("pageIndex must be a valid integer")
			}
			pageIndex = int32(pageIndex64)
		}

		var pageSize int32 = 30
		if pageSizeStr != "" {
			pageSize64, err := strconv.ParseInt(pageSizeStr, 10, 32)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("pageSize must be a valid integer")
			}
			pageSize = int32(pageSize64)
		}

		query := &queries.GetAllDeviceDefinitionOnChainQuery{MakeSlug: make, PageIndex: pageIndex, PageSize: pageSize}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
