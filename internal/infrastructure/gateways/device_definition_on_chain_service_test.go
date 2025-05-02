package gateways

import (
	_ "embed" // import the embed package
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"

	"github.com/stretchr/testify/assert"
)

func Test_deviceDefinitionOnChainService_validateAttributes(t *testing.T) {
	tests := []struct {
		name              string
		currentAttributes []models.DeviceTypeAttribute
		newAttributes     []models.DeviceTypeAttribute
		expectedNewOrMod  []models.DeviceTypeAttribute
		expectedRemoved   []models.DeviceTypeAttribute
	}{
		{
			name: "add new attribute",
			currentAttributes: []models.DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
			},
			newAttributes: []models.DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
				{Name: "number_of_doors", Value: "4"},
			},
			expectedNewOrMod: []models.DeviceTypeAttribute{
				{Name: "number_of_doors", Value: "4"},
			},
			expectedRemoved: []models.DeviceTypeAttribute{},
		},
		{
			name: "remove current attribute",
			currentAttributes: []models.DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
			},
			newAttributes: []models.DeviceTypeAttribute{
				{Name: "powertrain_type", Value: "BEV"},
			},
			expectedNewOrMod: []models.DeviceTypeAttribute{
				{Name: "powertrain_type", Value: "BEV"},
			},
			expectedRemoved: []models.DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
			},
		},
		{
			name: "update current attribute",
			currentAttributes: []models.DeviceTypeAttribute{
				{Name: "fuel_type", Value: "AWD"},
			},
			newAttributes: []models.DeviceTypeAttribute{
				{Name: "fuel_type", Value: "Electric"},
			},
			expectedNewOrMod: []models.DeviceTypeAttribute{
				{Name: "fuel_type", Value: "Electric"},
			},
			expectedRemoved: []models.DeviceTypeAttribute{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newOrModified, removed := validateAttributes(test.currentAttributes, test.newAttributes)

			assert.ElementsMatch(t, test.expectedNewOrMod, newOrModified, "New or modified attributes should match expected")
			assert.ElementsMatch(t, test.expectedRemoved, removed, "Removed attributes should match expected")
		})
	}
}
