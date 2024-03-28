//nolint:tagliatelle
package models

import (
	"time"
)

type GetIntegrationFeatureQueryResult struct {
	FeatureKey      string    `json:"feature_key"`
	ElasticProperty string    `json:"elastic_property"`
	DisplayName     string    `json:"display_name"`
	CSSIcon         string    `json:"css_icon,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	FeatureWeight   float64   `json:"feature_weight,omitempty"`
}

type UpdateDeviceIntegrationFeatureAttribute struct {
	FeatureKey   string `json:"featureKey"`
	SupportLevel int16  `json:"supportLevel"`
}
