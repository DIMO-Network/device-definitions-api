//nolint:tagliatelle
package models

import "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

type GetDeviceStyleQueryResult struct {
	ID                 string                              `json:"id"`
	DefinitionID       string                              `json:"definition_id"`
	DeviceDefinition   GetDeviceDefinitionStyleQueryResult `json:"device_definition"`
	Name               string                              `json:"name"`
	ExternalStyleID    string                              `json:"external_style_id"`
	Source             string                              `json:"source"`
	SubModel           string                              `json:"sub_model"`
	HardwareTemplateID string                              `json:"hardware_template_id"`
}

type GetDeviceDefinitionStyleQueryResult struct {
	DeviceAttributes []DeviceTypeAttribute `json:"deviceAttributes"`
}

func ConvertMetadataToDeviceAttributes(metadata *gateways.DeviceDefinitionMetadata) []DeviceTypeAttribute {
	// Depending on your types, you might have to perform additional conversion logic.
	// Here we're simply returning the metadata as-is for assignment, but if your
	// DeviceAttributes require a specific structure, you'll have to adjust this.
	dta := []DeviceTypeAttribute{}
	if metadata == nil {
		return dta
	}
	for _, attribute := range metadata.DeviceAttributes {
		dta = append(dta, DeviceTypeAttribute{
			Name:        attribute.Name,
			Label:       attribute.Name,
			Description: attribute.Name,
			Type:        "",
			Required:    false,
			Value:       attribute.Value,
			Option:      nil,
		})
	}
	return dta
}
