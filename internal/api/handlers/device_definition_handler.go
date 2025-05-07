package handlers

import (
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"

	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	_ "github.com/DIMO-Network/device-definitions-api/internal/core/models" // required for swagger to generate models
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	_ "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways" // required for swagger to generate models
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
			DeviceDefinitionID: resp.DefinitionId,
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
	// if a new device definition was created, the tableland transaction hash from the insert statement. Check this has completed before querying the ID
	NewTransactionHash string `json:"newTransactionHash"`
}

// GetDeviceDefinitionByID godoc
// @Summary gets a device definition, from tableland on-chain records. Only support mmy style id's eg. ford_escape_2025
// @ID GetDeviceDefinitionByID
// @Description gets a device definition
// @Tags device-definitions
// @Param  id path string true "mmy definition_id eg. ford_escape_2020"
// @Produce json
// @Success 200 {object} models.DeviceDefinitionTablelandModel
// @Failure 404
// @Failure 400
// @Failure 500
// @Router /device-definitions/{id} [get]
func GetDeviceDefinitionByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		// make sure it is mmy style dd id
		split := strings.Split(id, "_")
		if len(split) == 3 {
			query := &queries.GetDeviceDefinitionByIDQueryV2{DefinitionID: id}
			result, _ := m.Send(c.UserContext(), query)
			return c.Status(fiber.StatusOK).JSON(result)
		}

		return c.Status(fiber.StatusBadRequest).JSON(common.ProblemDetails{
			Type:   "https://tools.ietf.org/html/rfc7231#section-6.5.1",
			Title:  "invalid id format",
			Status: fiber.StatusBadRequest,
			Detail: "id must be mmy style eg. ford_escape_2025",
		})
	}
}

// VINProfile godoc
// @Summary gets any raw profile info we have on previously decoded VINs. USA Only.
// @ID VINProfile
// @Description gets VIN profile if we have it.
// @Tags device-definitions
// @Param  vin path string true "17 character usa based VIN eg. WBA12345678901234"
// @Produce json
// @Success 200 {object} queries.GetVINProfileResponse
// @Failure 404
// @Failure 400
// @Failure 500
// @Router /vin-profile/{vin} [get]
func VINProfile(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vin := c.Params("vin")
		// make sure it is mmy style dd id
		if len(vin) != 17 {
			return c.Status(fiber.StatusBadRequest).JSON(common.ProblemDetails{
				Type:   "https://tools.ietf.org/html/rfc7231#section-6.5.1",
				Title:  "invalid VIN format",
				Status: fiber.StatusBadRequest,
				Detail: "Only USA style VINs supported 17 characters long.",
			})
		}

		query := &queries.GetVINProfileQuery{VIN: vin}
		result, err := m.Send(c.UserContext(), query)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(common.ProblemDetails{
				Type:   "https://tools.ietf.org/html/rfc7231#section-6.5.1",
				Title:  "No VIN profile founder",
				Status: fiber.StatusNotFound,
				Detail: "Couldn't get VIN profile.",
			})
		}
		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetR1CompatibilitySearch godoc
// @Summary gets r1 MMY compatibility by search filter
// @ID GetR1CompatibilitySearch
// @Description gets r1 compatibility search by filter
// @Tags device-definitions
// @Param  query query string true "query filter"
// @Param  page query number false "page"
// @Param  pageSize query number false "pageSize"
// @Accept json
// @Produce json
// @Success 200 {object} queries.GetR1CompatibilitySearchQueryResult
// @Failure 500
// @Router /device-definitions/search-r1 [get]
func GetR1CompatibilitySearch(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		q := c.Query("query")

		defaultPage := 1
		defaultPageSize := 20

		page := c.Query("page", strconv.Itoa(defaultPage))
		pageSize := c.Query("pageSize", strconv.Itoa(defaultPageSize))

		pageInt, _ := strconv.Atoi(page)
		pageSizeInt, _ := strconv.Atoi(pageSize)

		query := &queries.GetR1CompatibilitySearch{Query: q, PageSize: pageSizeInt, Page: pageInt}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetCompatibilityR1Sheet godoc
// @Summary gets r1 MMY compatibility google sheet in JSON form
// @ID GetCompatibilityR1Sheet
// @Description gets r1 MMY compatibility google sheet in JSON form. returns an array of below objects
// @Tags device-definitions
// @Produce json
// @Success 200 {object} queries.CompatibilitySheetRow
// @Failure 500
// @Router /compatibility/r1-sheet [get]
func GetCompatibilityR1Sheet(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		query := &queries.GetCompatibilityR1SheetQuery{}
		result, _ := m.Send(c.UserContext(), query)
		c.Set("Cache-Control", "public, max-age=600, s-maxage=600, immutable")
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
