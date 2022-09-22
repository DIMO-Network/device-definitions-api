package models

type IntegrationFeatures struct {
	FeatureKey      string `json:"feature_key,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	IconCss         string `json:"css_icon,omitempty"`
	ElasticProperty string `json:"elastic_property,omitempty"`
}

type DeviceIntegrationFeatures struct {
	ElasticProperty string `json:"elastic_property,omitempty"`
	SupportLevel    int8   `json:"supportLevel"`
}
