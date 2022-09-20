package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

// GetDeviceDefinitionByID godoc
// @Summary gets a device definition
// @ID GetByID
// @Description gets a device definition
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetByIdQueryResult
// @Failure 404 {object} common.ProblemDetails{}
// @Failure 500 {object} common.ProblemDetails{}
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
// @ID GetDeviceDefinitionByMakeModelsAndYears
// @Description gets a device definition
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetAllQueryResult
// @Failure 500 {object} common.ProblemDetails{}
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
// @ID GetDeviceIntegrationsByID
// @Description gets a specific device definition by make model and year.
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetByModelYearQueryResult
// @Failure 404 {object} common.ProblemDetails{}
// @Failure 500 {object} common.ProblemDetails{}
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

func GetDeviceDefinitionByDynamicFilter(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		deviceDefinition := c.Query("deviceMakeId")
		integration := c.Query("integration_id")
		make := c.Query("make_id")
		model := c.Query("model")
		year := c.Query("year")
		yrInt, _ := strconv.Atoi(year)
		verifiedVinList := c.Query("verified_vin_list")
		pageIndex := c.Query("page_index")
		pgInInt, _ := strconv.Atoi(pageIndex)
		pageSize := c.Query("page_size")
		pgSzInt, _ := strconv.Atoi(pageSize)

		var verifiedVinArr []string
		_ = json.Unmarshal([]byte(verifiedVinList), &verifiedVinArr)

		query := &queries.GetDeviceDefinitionByDynamicFilterQuery{
			MakeID:             make,
			IntegrationID:      integration,
			DeviceDefinitionID: deviceDefinition,
			Year:               yrInt,
			Model:              model,
			VerifiedVinList:    verifiedVinArr,
			PageIndex:          pgInInt,
			PageSize:           pgSzInt,
		}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
