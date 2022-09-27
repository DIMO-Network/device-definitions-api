package commands

import (
	"context"
	_ "embed"
	"testing"

	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UpdateDeviceDefinitionCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	pdb                       db.Store
	container                 testcontainers.Container
	ctx                       context.Context
	mockDeviceDefinitionCache *mockService.MockDeviceDefinitionCacheService

	commandHandler UpdateDeviceDefinitionCommandHandler
}

func TestUpdateDeviceDefinitionCommandHandler(t *testing.T) {
	suite.Run(t, new(UpdateDeviceDefinitionCommandHandlerSuite))
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.commandHandler = NewUpdateDeviceDefinitionCommandHandler(s.pdb.DBS, s.mockDeviceDefinitionCache)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)

	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		VehicleInfo: UpdateDeviceVehicleInfo{
			FuelType:            "test",
			DrivenWheels:        "test",
			NumberOfDoors:       "4",
			BaseMSRP:            1,
			EPAClass:            "test",
			VehicleType:         "test",
			MPGHighway:          "1",
			MPGCity:             "1",
			FuelTankCapacityGal: "1",
			MPG:                 "1",
		},
	})
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_WithNewStyles_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)

	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	var deviceStyles []UpdateDeviceStyles
	deviceStyles = append(deviceStyles, UpdateDeviceStyles{
		ID:              ksuid.New().String(),
		Name:            "NewStyle1",
		Source:          "Source",
		SubModel:        "SubModel",
		ExternalStyleID: ksuid.New().String(),
	})
	deviceStyles = append(deviceStyles, UpdateDeviceStyles{
		ID:              ksuid.New().String(),
		Name:            "NewStyle2",
		Source:          "Source",
		SubModel:        "SubModel2",
		ExternalStyleID: ksuid.New().String(),
	})

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		VehicleInfo: UpdateDeviceVehicleInfo{
			FuelType:            "test",
			DrivenWheels:        "test",
			NumberOfDoors:       "4",
			BaseMSRP:            1,
			EPAClass:            "test",
			VehicleType:         "test",
			MPGHighway:          "1",
			MPGCity:             "1",
			FuelTankCapacityGal: "1",
			MPG:                 "1",
		},
		DeviceStyles: deviceStyles,
	})
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)

	dd, _ = models.DeviceDefinitions(
		qm.Where("id = ?", dd.ID),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceStyles))).
		One(ctx, s.pdb.DBS().Writer)

	assert.Equal(s.T(), len(dd.R.DeviceStyles), 2)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_WithNewIntegration_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	i := setupIntegrationForSmartCarCompatibility(s.T(), s.pdb)
	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)

	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	var deviceIntegrations []UpdateDeviceIntegrations
	deviceIntegrations = append(deviceIntegrations, UpdateDeviceIntegrations{
		IntegrationID: i.ID,
		Region:        "us-01",
	})

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		VehicleInfo: UpdateDeviceVehicleInfo{
			FuelType:            "test",
			DrivenWheels:        "test",
			NumberOfDoors:       "4",
			BaseMSRP:            1,
			EPAClass:            "test",
			VehicleType:         "test",
			MPGHighway:          "1",
			MPGCity:             "1",
			FuelTankCapacityGal: "1",
			MPG:                 "1",
		},
		DeviceIntegrations: deviceIntegrations,
	})
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)

	dd, _ = models.DeviceDefinitions(
		qm.Where("id = ?", dd.ID),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration))).
		One(ctx, s.pdb.DBS().Writer)

	assert.Equal(s.T(), len(dd.R.DeviceIntegrations), 1)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_Exception() {
	ctx := context.Background()

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: "dd.ID",
	})

	s.Nil(commandResult)
	s.Error(err)
}

func setupDeviceDefinitionForUpdate(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}
