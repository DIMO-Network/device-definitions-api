//nolint:tagliatelle
package models

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
	DeviceAttributes []DeviceTypeAttributeEditor `json:"deviceAttributes"`
}
