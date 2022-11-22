package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

func Test_getIntegrationFeatures(t *testing.T) {
	ctx := context.Background()
	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
		teslaMakeID          = "2681caeN3FuuACJ819ORd1YLvEZ"
	)
	pdb, container := dbtesthelper.StartContainerDatabase(ctx, dbName, t, migrationsDirRelPath)
	defer container.Terminate(ctx)                      //nolint
	dbtesthelper.TruncateTables(pdb.DBS().Writer.DB, t) // clear setup data for integration features
	// arrange some data
	feat1 := models.IntegrationFeature{
		FeatureKey:      "tires",
		ElasticProperty: "tires",
		DisplayName:     "Tires",
		FeatureWeight:   null.Float64From(1.0),
		PowertrainType:  models.PowertrainALL,
	}
	err := feat1.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)
	feat2 := models.IntegrationFeature{
		FeatureKey:      "odometer",
		ElasticProperty: "odometer",
		DisplayName:     "Odometer",
		FeatureWeight:   null.Float64From(0.75),
		PowertrainType:  models.PowertrainBEV,
	}
	err = feat2.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)
	// this one should be ignored
	feat3 := models.IntegrationFeature{
		FeatureKey:      "fuelTankCapacity",
		ElasticProperty: "fuelTankCapacity",
		DisplayName:     "FuelTankCapacity",
		FeatureWeight:   null.Float64From(0.75),
		PowertrainType:  models.PowertrainHybridsAndICE,
	}
	err = feat3.Insert(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	features, totalWeights, err := getIntegrationFeatures(ctx, teslaMakeID, pdb.DBS().Reader)
	require.NoError(t, err)

	assert.Equal(t, 1.75, totalWeights)
	assert.Len(t, features, 2)
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
