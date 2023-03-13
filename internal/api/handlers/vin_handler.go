package handlers

import (
	"encoding/csv"
	"fmt"

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
// @Param  vins body []string true "vin array."
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Router /bulk-decode [post]
func BulkDecodeVIN(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		payload := make([]string, 0)

		err := c.BodyParser(&payload)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}

		command := &commands.BulkValidateVinCommand{VINs: payload}

		result, _ := m.Send(c.UserContext(), command)

		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// BulkDecodeVIN CSV godoc
// @Summary gets a list of decoded vins in csv format.
// @ID BulkDecodeVINCSV
// @Description decodes a list of VINs
// @Tags device-definitions
// @Accept json
// @Param  vins body []string true "vin array."
// @Produce text/csv
// @Success 200
// @Failure 404
// @Failure 500
// @Router /bulk-decode/csv [post]
func BulkDecodeVINCsv(m mediator.Mediator) fiber.Handler {
	return func(c *fiber.Ctx) error {

		payload := make([]string, 0)

		err := c.BodyParser(&payload)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(err)
		}

		command := &commands.BulkValidateVinCommand{VINs: payload}

		result, _ := m.Send(c.UserContext(), command)

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=bulk-decode.csv")

		decodedVINs := result.(commands.BulkValidateVinCommandResult).DecodedVINs

		writer := csv.NewWriter(c)

		defer writer.Flush()

		header := []string{
			"VIN",
			"Device Definition ID",
			"Device Make ID",
			"Device Make Name",
			"Device Model",
			"Device Year",
			"Compatibility Feature Integration ID",
			"Compatibility Feature Integration Vendor",
			"Compatibility Feature Region",
			"Compatibility Feature Level",
		}

		err = writer.Write(header)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err)
		}

		for _, v := range decodedVINs {

			row := []string{
				v.VIN,
				v.DeviceDefinitionID,
				v.DeviceMake.ID,
				v.DeviceMake.Name,
				v.DeviceModel,
				fmt.Sprint(v.DeviceYear),
			}

			for _, f := range v.CompatibilityFeatures {
				row = append(row, f.IntegrationId)
				row = append(row, f.IntegrationVendor)
				row = append(row, f.Region)
				row = append(row, f.Level)
			}

			err := writer.Write(row)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(err)
			}
		}

		return nil
	}
}
