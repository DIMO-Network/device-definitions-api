package gateways

import (
	_ "embed" // import the embed package
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_deviceDefinitionOnChainService_validateAttributes(t *testing.T) {
	tests := []struct {
		name              string
		currentAttributes []DeviceTypeAttribute
		newAttributes     []DeviceTypeAttribute
		expectedNewOrMod  []DeviceTypeAttribute
		expectedRemoved   []DeviceTypeAttribute
	}{
		{
			name: "add new attribute",
			currentAttributes: []DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
			},
			newAttributes: []DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
				{Name: "number_of_doors", Value: "4"},
			},
			expectedNewOrMod: []DeviceTypeAttribute{
				{Name: "number_of_doors", Value: "4"},
			},
			expectedRemoved: []DeviceTypeAttribute{},
		},
		{
			name: "remove current attribute",
			currentAttributes: []DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
			},
			newAttributes: []DeviceTypeAttribute{
				{Name: "powertrain_type", Value: "BEV"},
			},
			expectedNewOrMod: []DeviceTypeAttribute{
				{Name: "powertrain_type", Value: "BEV"},
			},
			expectedRemoved: []DeviceTypeAttribute{
				{Name: "driven_wheels", Value: "AWD"},
				{Name: "manufacturer_code", Value: "K3S"},
			},
		},
		{
			name: "update current attribute",
			currentAttributes: []DeviceTypeAttribute{
				{Name: "fuel_type", Value: "AWD"},
			},
			newAttributes: []DeviceTypeAttribute{
				{Name: "fuel_type", Value: "Electric"},
			},
			expectedNewOrMod: []DeviceTypeAttribute{
				{Name: "fuel_type", Value: "Electric"},
			},
			expectedRemoved: []DeviceTypeAttribute{},
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
