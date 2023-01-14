package main

import (
	"context"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	elasticModels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch/models"
	"github.com/rs/zerolog"
	"os"
	"testing"
)

const (
	dbName               = "device_definitions_api"
	migrationsDirRelPath = "../../internal/infrastructure/db/migrations"
)

func TestcopyFeaturesToRegion(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	regionToFeats := map[string][]elasticModels.DeviceIntegrationFeatures{}
	pdb, container := dbtesthelper.StartContainerDatabase(ctx, dbName, t, migrationsDirRelPath)
	defer container.Stop(ctx, nil) // nolint
	autopiInt := dbtesthelper.SetupCreateAutoPiIntegration(t, pdb)
	_ = dbtesthelper.SetupCreateDeviceType(t, pdb)
	dm := dbtesthelper.SetupCreateMake(t, "Ford", pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, "Mach-E", 2022, pdb)
	di1 := dbtesthelper.SetupCreateDeviceIntegration(t, dd, autopiInt.ID, pdb)
	// todo need another like above but for Europe

	// todo add stuff to regionToFeatures
	//act
	copyFeaturesToRegion(ctx, logger, regionToFeats, pdb, dd.ID, autopiInt.ID)

	// assert via DB query
}
