package models

type GetDeviceTypeQueryResult struct {
	ID          string                              `json:"id"`
	Name        string                              `json:"name"`
	Metadatakey string                              `json:"metadata_key"`
	Attributes  []GetDeviceTypeAttributeQueryResult `json:"attributes"`
}

type GetDeviceTypeAttributeQueryResult struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	DefaultValue string   `json:"default_value"`
	Option       []string `json:"options"`
}

type CreateDeviceTypeAttribute struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	DefaultValue string   `json:"default_value"`
	Option       []string `json:"options"`
}

type UpdateDeviceTypeAttribute struct {
	// Name should match one of the name keys in the allowed device_types.properties
	Name  string `json:"name"`
	Value string `json:"value"`
}
