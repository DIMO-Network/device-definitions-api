package models

type IntegrationFeatures struct {
	FeatureKey      string `json:"feature_key,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	IconCSS         string `json:"css_icon,omitempty"`
	ElasticProperty string `json:"elastic_property,omitempty"`
}

type DeviceIntegrationFeatures struct {
	FeatureKey   string `json:"feature_key,omitempty"`
	SupportLevel int8   `json:"support_level"`
}
