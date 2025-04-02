package handlers

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/gofiber/fiber/v2"
)

// GetIntegrations godoc
// @Summary gets list of integrations we have defined.
// @ID GetIntegrations
// @Description gets list of integrations we have defined.
// @Tags device-definitions
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /integrations [get]
func GetIntegrations(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		query := &queries.GetAllIntegrationQuery{}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// GetIntegrationByID godoc
// @Summary gets integration by id.
// @ID GetIntegrationByID
// @Description gets integration by id.
// @Tags device-definitions
// @Accept json
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /integrations/{id} [get]
func GetIntegrationByID(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		query := &queries.GetIntegrationByIDQuery{IntegrationID: []string{id}}

		result, _ := m.Send(c.UserContext(), query)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

//go:embed smartcar_oems.json
var smartcarOems []byte

// GetSmartcarManufacturers godoc
// @Summary gets all supported manufacturers for the smartcar external integration
// @ID GetSmartcarManufacturers
// @Description gets manufacturers supported by smartcar
// @Tags device-definitions
// @Produce json
// @Success 200
// @Failure 500
// @Router /manufacturers/integrations/smartcar [get]
func GetSmartcarManufacturers() fiber.Handler {
	const assetAPI = "https://api.dimo.co/assets/logos/"

	return func(c *fiber.Ctx) error {
		var jsonContent []map[string]interface{}
		if err := json.Unmarshal(smartcarOems, &jsonContent); err != nil {
			// If there's an error parsing the JSON, return a 400 status
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Invalid JSON format: %v", err))
		}
		// Prepend the url path to each "logo" field in the array
		for _, item := range jsonContent {
			if logo, ok := item["logo"].(string); ok {
				item["logo"] = assetAPI + logo
			}
		}

		return c.JSON(jsonContent)
	}
}
