package models

type GetDeviceTypeQueryResult struct {
	ID         string                              `json:"id"`
	Name       string                              `json:"name"`
	Attributes []GetDeviceTypeAttributeQueryResult `json:"attributes"`
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
