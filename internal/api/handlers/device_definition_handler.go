package handlers

import (
	"strconv"

	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	_ "github.com/DIMO-Network/device-definitions-api/internal/core/models" // required for swagger to generate modesl
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/gofiber/fiber/v2"
)

// DecodeVIN godoc
// @Summary returns device definition id corresponding to a given vin
// @ID DecodeVIN
// @Description decode a vin
// @Tags device-definitions
// @Produce json
// @Accept json
// @Param  decodeRequest body DecodeVINRequest true  "Decode VIN request"
// @Success 200 {object} DecodeVINResponse "Response with definition ID.
// @Failure 404
// @Failure 500
// @Security    BearerAuth
// @Router /device-definitions/decode-vin [post]
func DecodeVIN(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var dvr DecodeVINRequest
		if err := c.BodyParser(&dvr); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Couldn't parse request body.")
		}
		query := &queries.DecodeVINQuery{VIN: dvr.VIN, Country: dvr.CountryCode}

		result, err := m.Send(c.UserContext(), query)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Couldn't decode VIN.")
		}
		resp := result.(*p_grpc.DecodeVinResponse)
		dd := DecodeVINResponse{
			DeviceDefinitionID: resp.NameSlug,
			LegacyID:           resp.DeviceDefinitionId,
			NewTransactionHash: resp.NewTrxHash,
		}

		return c.Status(fiber.StatusOK).JSON(dd)
	}
}

type DecodeVINRequest struct {
	VIN string `json:"vin"`
	// 3 letter ISO
	CountryCode string `json:"countryCode"`
}

type DecodeVINResponse struct {
	// new name slug based id, can use this to query identity-api
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	// old ksuid based device def id
	LegacyID string `json:"legacyId"`
	// if a new device definition was created, the tableland transaction hash from the insert statement. Check this has completed before querying the ID
	NewTransactionHash string `json:"newTransactionHash"`
}

// GetDeviceDefinitionByID godoc
// @Summary gets a device definition
// @ID GetDeviceDefinitionByID
// @Description gets a device definition
// @Tags device-definitions
// @Param  id path string true "device definition id"
// @Produce json
// @Success 200 {object} models.GetDeviceDefinitionQueryResult
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
// @Param  make query string true "make"
// @Param  model query string true "model"
// @Param  year query number true "year"
// @Produce json
// @Success 200 {object} models.GetDeviceDefinitionQueryResult
// @Failure 404
// @Failure 500
// @Router /device-definitions [get]
func GetDeviceDefinitionByMMY(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mk := c.Query("make")
		model := c.Query("model")
		year := c.Query("year")
		yrInt, _ := strconv.Atoi(year)

		query := &queries.GetDeviceDefinitionByMakeModelYearQuery{Make: mk, Model: model, Year: yrInt}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetDeviceDefinitionSearch godoc
// @Summary gets device definitions by search filter
// @ID GetDeviceDefinitionSearch
// @Description gets a device definition by search filter
// @Tags device-definitions
// @Param  query query string true "query filter"
// @Param  makeSlug query string false "make Slug"
// @Param  modelSlug query string false "model Slug"
// @Param  year query number false "year"
// @Param  page query number false "page"
// @Param  pageSize query number false "pageSize"
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetAllDeviceDefinitionBySearchQueryResult
// @Failure 500
// @Router /device-definitions/search [get]
func GetDeviceDefinitionSearch(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		q := c.Query("query")
		mk := c.Query("makeSlug")
		model := c.Query("modelSlug")
		year := c.Query("year")
		yrInt, _ := strconv.Atoi(year)

		defaultPage := 1
		defaultPageSize := 20

		page := c.Query("page", strconv.Itoa(defaultPage))
		pageSize := c.Query("pageSize", strconv.Itoa(defaultPageSize))

		pageInt, _ := strconv.Atoi(page)
		pageSizeInt, _ := strconv.Atoi(pageSize)

		query := &queries.GetAllDeviceDefinitionBySearchQuery{Query: q, Make: mk, Model: model, Year: yrInt, PageSize: pageSizeInt, Page: pageInt}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetDeviceDefinitionAutocomplete godoc
// @Summary gets device definitions autocomplete
// @ID GetDeviceDefinitionAutocomplete
// @Description gets a device definition Autocomplete
// @Tags device-definitions
// @Param  query query string true "query filter"
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetAllDeviceDefinitionByAutocompleteQueryResult
// @Failure 500
// @Router /device-definitions/autocomplete [get]
func GetDeviceDefinitionAutocomplete(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		q := c.Query("query")

		query := &queries.GetAllDeviceDefinitionByAutocompleteQuery{Query: q}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}
