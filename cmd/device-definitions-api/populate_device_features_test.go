package main

import (
	"context"
	"os"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	elasticModels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dbName               = "device_definitions_api"
	migrationsDirRelPath = "../../internal/infrastructure/db/migrations"
)

func Test_copyFeaturesToRegion(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	regionToFeats := map[string][]elasticModels.DeviceIntegrationFeatures{}
	pdb, container := dbtesthelper.StartContainerDatabase(ctx, dbName, t, migrationsDirRelPath)
	defer container.Stop(ctx, nil) // nolint
	autopiInt := dbtesthelper.SetupCreateAutoPiIntegration(t, pdb)
	_ = dbtesthelper.SetupCreateDeviceType(t, pdb)
	dm := dbtesthelper.SetupCreateMake(t, "Ford", pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, "Mach-E", 2022, pdb)
	_ = dbtesthelper.SetupCreateDeviceIntegration(t, dd, autopiInt.ID, "Americas", pdb)
	_ = dbtesthelper.SetupCreateDeviceIntegration(t, dd, autopiInt.ID, "Europe", pdb)
	_ = dbtesthelper.SetupCreateDeviceIntegration(t, dd, autopiInt.ID, "Asia", pdb)

	regionToFeats["Americas"] = []elasticModels.DeviceIntegrationFeatures{{
		FeatureKey:   "odometer",
		SupportLevel: Supported.Int(),
	}, {
		FeatureKey:   "tires",
		SupportLevel: Supported.Int(),
	}, {
		FeatureKey:   "oil",
		SupportLevel: NotSupported.Int(),
	}}
	//act
	copyFeaturesToMissingRegion(ctx, logger, regionToFeats, pdb, dd.ID, autopiInt.ID)
	// assert via DB query
	updatedEurope, err := models.FindDeviceIntegration(ctx, pdb.DBS().Reader, dd.ID, autopiInt.ID, "Europe")
	require.NoError(t, err)
	var newFeats []elasticModels.DeviceIntegrationFeatures

	err = updatedEurope.Features.Unmarshal(&newFeats)
	require.NoError(t, err)
	require.Len(t, newFeats, 3)
	assert.Equal(t, "odometer", newFeats[0].FeatureKey)
	assert.Equal(t, "tires", newFeats[1].FeatureKey)
	assert.Equal(t, MaybeSupported.Int(), newFeats[0].SupportLevel)
	assert.Equal(t, MaybeSupported.Int(), newFeats[1].SupportLevel)

	assert.Equal(t, "oil", newFeats[2].FeatureKey)
	assert.Equal(t, NotSupported.Int(), newFeats[2].SupportLevel)

	// -- validate for third Region, just check length since already validated rest above
	updatedAsia, err := models.FindDeviceIntegration(ctx, pdb.DBS().Reader, dd.ID, autopiInt.ID, "Asia")
	require.NoError(t, err)
	err = updatedAsia.Features.Unmarshal(&newFeats)
	require.NoError(t, err)
	require.Len(t, newFeats, 3)
	assert.Equal(t, "odometer", newFeats[0].FeatureKey)
	assert.Equal(t, MaybeSupported.Int(), newFeats[0].SupportLevel)
}
