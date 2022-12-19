package models

type GetDeviceStyleQueryResult struct {
	ID                 string `json:"id"`
	DeviceDefinitionID string `json:"device_definition_id"`
	Name               string `json:"name"`
	ExternalStyleID    string `json:"external_style_id"`
	Source             string `json:"source"`
	SubModel           string `json:"sub_model"`
	HardwareTemplateID string `json:"hardware_template_id"`
}
