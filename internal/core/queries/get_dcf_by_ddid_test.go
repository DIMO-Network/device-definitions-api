package queries

import (
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"

	"testing"
)

func Test_buildFeatures(t *testing.T) {
	json := null.JSONFrom([]byte(deviceIntegrationFeaturesJSON))

	fs := buildFeatures(json, integrationFeatures)

	assert.Len(t, fs, 4)
	assert.Equal(t, int32(0), findFeat("tires", fs).SupportLevel)
	assert.Equal(t, "tires", findFeat("tires", fs).CssIcon)
	assert.Equal(t, "tires", findFeat("tires", fs).Key)
	assert.Equal(t, "Tires", findFeat("tires", fs).DisplayName)

	assert.Equal(t, int32(2), findFeat("cell_tower", fs).SupportLevel)
	assert.Equal(t, "cell_tower", findFeat("cell_tower", fs).CssIcon)
	assert.Equal(t, "cell_tower", findFeat("cell_tower", fs).Key)
	assert.Equal(t, "Cell Tower", findFeat("cell_tower", fs).DisplayName)
	// support level should be 0 b/c not in the JSON
	assert.Equal(t, int32(0), findFeat("something_else", fs).SupportLevel)
	assert.Equal(t, "something_else", findFeat("something_else", fs).Key)
	assert.Equal(t, "Something Else", findFeat("something_else", fs).DisplayName)
}

func Test_buildFeatures_noData(t *testing.T) {
	json := null.JSON{}

	fs := buildFeatures(json, integrationFeatures)

	assert.Nil(t, fs)
}

func Test_calculateCompatibilityLevel(t *testing.T) {
	json := null.JSONFrom([]byte(deviceIntegrationFeaturesJSON))

	fs := buildFeatures(json, integrationFeatures)

	level := calculateCompatibilityLevel(fs, integrationFeatures, 2.0)
	// 50%
	assert.Equal(t, SilverLevel, level)
}

func findFeat(key string, fs []*p_grpc.Feature) *p_grpc.Feature {
	for _, f := range fs {
		if f.Key == key {
			return f
		}
	}
	return nil
}

var integrationFeatures = models.IntegrationFeatureSlice{
	&models.IntegrationFeature{
		FeatureKey:    "tires",
		DisplayName:   "Tires",
		CSSIcon:       null.StringFrom("tires"),
		FeatureWeight: null.Float64From(0.50),
	},
	&models.IntegrationFeature{
		FeatureKey:    "cell_tower",
		DisplayName:   "Cell Tower",
		CSSIcon:       null.StringFrom("cell_tower"),
		FeatureWeight: null.Float64From(0.50),
	},
	// css icon not set
	&models.IntegrationFeature{
		FeatureKey:    "engine_speed",
		DisplayName:   "Engine Speed",
		FeatureWeight: null.Float64From(0.50),
	},
	// feature that does not exist in json
	&models.IntegrationFeature{
		FeatureKey:    "something_else",
		DisplayName:   "Something Else",
		FeatureWeight: null.Float64From(0.50),
	},
}

var deviceIntegrationFeaturesJSON = `
	[{"featureKey": "tires", "supportLevel": 0}, 
{"featureKey": "vin", "supportLevel": 0}, 
{"featureKey": "cell_tower", "supportLevel": 2}, 
{"featureKey": "coolant_temperature", "supportLevel": 0}, 
{"featureKey": "engine_speed", "supportLevel": 2}, 
{"featureKey": "speed", "supportLevel": 2}, 
{"featureKey": "barometric_pressure", "supportLevel": 2}, 
{"featureKey": "oil", "supportLevel": 0},
{"featureKey": "ambient_temperature", "supportLevel": 2}, 
{"featureKey": "location", "supportLevel": 2}, 
{"featureKey": "odometer", "supportLevel": 0}, 
{"featureKey": "throttle_position", "supportLevel": 2}, 
{"featureKey": "engine_runtime", "supportLevel": 0}, 
{"featureKey": "ev_battery", "supportLevel": 0}, 
{"featureKey": "fuel_tank", "supportLevel": 2}, 
{"featureKey": "fuel_type", "supportLevel": 0}, 
{"featureKey": "battery_capacity", "supportLevel": 0}, 
{"featureKey": "battery_voltage", "supportLevel": 2}, 
{"featureKey": "charging", "supportLevel": 0}, 
{"featureKey": "engine_load", "supportLevel": 0}, 
{"featureKey": "range", "supportLevel": 0}]
`
