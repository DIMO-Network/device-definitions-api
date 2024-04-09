package commands

import (
	"context"
	_ "embed"
	"testing"

	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

type UpdateDeviceDefinitionCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                               *gomock.Controller
	pdb                                db.Store
	container                          testcontainers.Container
	ctx                                context.Context
	mockDeviceDefinitionCache          *mockService.MockDeviceDefinitionCacheService
	mockRepository                     *repositoryMock.MockDeviceDefinitionRepository
	mockDeviceDefinitionOnChainService *mock_gateways.MockDeviceDefinitionOnChainService
	commandHandler                     UpdateDeviceDefinitionCommandHandler
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
	s.mockDeviceDefinitionOnChainService = mock_gateways.NewMockDeviceDefinitionOnChainService(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)
	s.mockRepository = repositoryMock.NewMockDeviceDefinitionRepository(s.ctrl)
	s.commandHandler = NewUpdateDeviceDefinitionCommandHandler(s.mockRepository, s.pdb.DBS, s.mockDeviceDefinitionCache)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_MetadataAttributes_withDB() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020
	// using real DB for integration test
	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)
	repo := repositories.NewDeviceDefinitionRepository(s.pdb.DBS, s.mockDeviceDefinitionOnChainService)
	cmdHandler := NewUpdateDeviceDefinitionCommandHandler(repo, s.pdb.DBS, s.mockDeviceDefinitionCache)

	trx_hash_hex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().CreateOrUpdate(ctx, gomock.Any(), gomock.Any()).Return(&trx_hash_hex, nil)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheBySlug(ctx, gomock.Any(), gomock.Any()).Times(1)

	// should work with only ddId and DeviceAttributes
	commandResult, err := cmdHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		DeviceAttributes: []*coremodels.UpdateDeviceTypeAttribute{
			{Name: "fuel_type", Value: "test"},
			{Name: "driven_wheels", Value: "test"},
			{Name: "number_of_doors", Value: "4"},
			{Name: "base_msrp", Value: "1"},
			{Name: "epa_class", Value: "test"},
			{Name: "mpg_highway", Value: "1"},
			{Name: "mpg_city", Value: "1"},
			{Name: "fuel_tank_capacity_gal", Value: "1"},
			{Name: "mpg", Value: "1"},
		},
	})
	s.Require().NoError(err)
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	assert.Equal(s.T(), result.ID, dd.ID)

	updatedDef, err := models.FindDeviceDefinition(ctx, s.pdb.DBS().Reader, dd.ID)
	s.Require().NoError(err)
	assert.Equal(s.T(), true, updatedDef.Metadata.Valid)
	j := updatedDef.Metadata.JSON
	assert.Equal(s.T(), "test", gjson.GetBytes(j, "vehicle_info.fuel_type").String())
	assert.Equal(s.T(), "test", gjson.GetBytes(j, "vehicle_info.driven_wheels").String())
	assert.Equal(s.T(), "4", gjson.GetBytes(j, "vehicle_info.number_of_doors").String())
	assert.Equal(s.T(), "1", gjson.GetBytes(j, "vehicle_info.base_msrp").String())
	assert.Equal(s.T(), "test", gjson.GetBytes(j, "vehicle_info.epa_class").String())
	assert.Equal(s.T(), "1", gjson.GetBytes(j, "vehicle_info.mpg_highway").String())
	assert.Equal(s.T(), "1", gjson.GetBytes(j, "vehicle_info.mpg_city").String())
	assert.Equal(s.T(), "1", gjson.GetBytes(j, "vehicle_info.mpg").String())
	assert.Equal(s.T(), "1", gjson.GetBytes(j, "vehicle_info.fuel_tank_capacity_gal").String())
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)

	s.mockRepository.EXPECT().GetByID(gomock.Any(), dd.ID).Return(dd, nil).AnyTimes()
	s.mockRepository.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dd, nil).AnyTimes()
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheBySlug(ctx, gomock.Any(), gomock.Any()).Times(1)

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		DeviceMakeID:       dd.R.DeviceMake.ID,
		Model:              model,
		Year:               int16(year),
		DeviceTypeID:       "vehicle",
		DeviceAttributes: []*coremodels.UpdateDeviceTypeAttribute{
			{Name: "fuel_type", Value: "test"},
			{Name: "driven_wheels", Value: "test"},
			{Name: "number_of_doors", Value: "4"},
			{Name: "base_msrp", Value: "1"},
			{Name: "epa_class", Value: "test"},
			{Name: "mpg_highway", Value: "1"},
			{Name: "mpg_city", Value: "1"},
			{Name: "fuel_tank_capacity_gal", Value: "1"},
			{Name: "mpg", Value: "1"},
		},
	})
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.Require().NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_WithNewStyles_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)
	i := setupNewIntegrationForUpdate(s.T(), s.pdb)

	s.mockRepository.EXPECT().GetByID(gomock.Any(), dd.ID).Return(dd, nil).AnyTimes()
	s.mockRepository.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dd, nil).AnyTimes()
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheBySlug(ctx, gomock.Any(), gomock.Any()).Times(1)

	deviceStyles := make([]*UpdateDeviceStyles, 2)
	deviceStyles[0] = &UpdateDeviceStyles{
		ID:              ksuid.New().String(),
		Name:            "NewStyle1",
		Source:          "Source",
		SubModel:        "SubModel",
		ExternalStyleID: ksuid.New().String(),
	}
	deviceStyles[1] = &UpdateDeviceStyles{
		ID:              ksuid.New().String(),
		Name:            "NewStyle2",
		Source:          "Source",
		SubModel:        "SubModel2",
		ExternalStyleID: ksuid.New().String(),
	}

	styles, _ := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID)).
		All(ctx, s.pdb.DBS().Reader)

	for _, style := range styles {
		deviceStyles = append(deviceStyles, &UpdateDeviceStyles{
			ID:              style.ID,
			Name:            style.Name,
			Source:          style.Source,
			SubModel:        style.SubModel,
			ExternalStyleID: ksuid.New().String(),
		})
	}

	integrations, _ := models.DeviceIntegrations(models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(dd.ID)).
		All(ctx, s.pdb.DBS().Reader)

	deviceIntegrations := make([]*UpdateDeviceIntegrations, len(integrations)+1)
	deviceIntegrations[0] = &UpdateDeviceIntegrations{
		IntegrationID: i.ID,
		Region:        "China",
	}

	for c, integration := range integrations {
		deviceIntegrations[c+1] = &UpdateDeviceIntegrations{
			IntegrationID: integration.IntegrationID,
			Region:        integration.Region,
			CreatedAt:     integration.CreatedAt,
		}
	}

	command := &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		Year:               2023,
		Model:              "M5",
		DeviceMakeID:       dd.DeviceMakeID,
		DeviceTypeID:       "vehicle",
		DeviceAttributes: []*coremodels.UpdateDeviceTypeAttribute{
			{Name: "fuel_type", Value: "test"},
			{Name: "driven_wheels", Value: "test"},
			{Name: "number_of_doors", Value: "4"},
			{Name: "base_msrp", Value: "1"},
			{Name: "epa_class", Value: "test"},
			{Name: "mpg_highway", Value: "1"},
			{Name: "mpg_city", Value: "1"},
			{Name: "fuel_tank_capacity_gal", Value: "1"},
			{Name: "mpg", Value: "1"},
		},
		DeviceStyles:       deviceStyles,
		DeviceIntegrations: deviceIntegrations,
	}

	commandResult, err := s.commandHandler.Handle(ctx, command)

	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.Require().NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)

}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_WithNewIntegration_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	i := setupIntegrationForSmartCarCompatibility(s.T(), s.pdb)
	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)
	dm := dbtesthelper.SetupCreateMake(s.T(), "BMW2", s.pdb)

	s.mockRepository.EXPECT().GetByID(gomock.Any(), dd.ID).Return(dd, nil).AnyTimes()
	s.mockRepository.EXPECT().CreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dd, nil).AnyTimes()

	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheBySlug(ctx, gomock.Any(), gomock.Any()).Times(1)

	var deviceIntegrations []*UpdateDeviceIntegrations
	deviceIntegrations = append(deviceIntegrations, &UpdateDeviceIntegrations{
		IntegrationID: i.ID,
		Region:        "us-01",
	})

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: dd.ID,
		Year:               2111,
		Model:              "M5",
		DeviceMakeID:       dm.ID,
		Verified:           true,
		DeviceTypeID:       "vehicle",
		DeviceAttributes: []*coremodels.UpdateDeviceTypeAttribute{
			{Name: "fuel_type", Value: "test"},
			{Name: "driven_wheels", Value: "test"},
			{Name: "number_of_doors", Value: "4"},
			{Name: "base_msrp", Value: "1"},
			{Name: "epa_class", Value: "test"},
			{Name: "mpg_highway", Value: "1"},
			{Name: "mpg_city", Value: "1"},
			{Name: "fuel_tank_capacity_gal", Value: "1"},
			{Name: "mpg", Value: "1"},
		},
		DeviceIntegrations: deviceIntegrations,
	})
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.Require().NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)
}

func (s *UpdateDeviceDefinitionCommandHandlerSuite) TestUpdateDeviceDefinitionCommand_Exception() {
	ctx := context.Background()

	s.mockRepository.EXPECT().GetByID(gomock.Any(), "dd.ID").Return(nil, errors.New("unhandled exception error")).AnyTimes()
	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: "dd.ID",
	})

	s.Nil(commandResult)
	s.Error(err)
}

func setupDeviceDefinitionForUpdate(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	_ = dbtesthelper.SetupCreateStyle(t, dd.ID, "4dr SUV 4WD", "edmunds", "Wagon", pdb)
	_ = dbtesthelper.SetupCreateStyle(t, dd.ID, "Hard Top 2dr SUV AWD", "edmunds", "Open Top", pdb)

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm

	return dd
}

func setupNewIntegrationForUpdate(t *testing.T, pdb db.Store) *models.Integration {
	i := dbtesthelper.SetupCreateHardwareIntegration(t, pdb)

	return i
}
