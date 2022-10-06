package models

type DeviceIntegrationFeatures struct {
	FeatureKey   string `json:"feature_key,omitempty"`
	SupportLevel int8   `json:"support_level"`
}
