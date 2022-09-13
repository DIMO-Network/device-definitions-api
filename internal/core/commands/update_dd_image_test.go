package commands

import (
	"context"
	_ "embed"
	"testing"

	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type UpdateDeviceDefinitionImageCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	pdb                       db.Store
	container                 testcontainers.Container
	ctx                       context.Context
	mockDeviceDefinitionCache *mockService.MockDeviceDefinitionCacheService

	commandHandler UpdateDeviceDefinitionImageCommandHandler
}

func TestUpdateDeviceDefinitionImageCommandHandler(t *testing.T) {
	suite.Run(t, new(UpdateDeviceDefinitionImageCommandHandlerSuite))
}

func (s *UpdateDeviceDefinitionImageCommandHandlerSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.commandHandler = NewUpdateDeviceDefinitionImageCommandHandler(s.pdb.DBS, s.mockDeviceDefinitionCache)
}

func (s *UpdateDeviceDefinitionImageCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *UpdateDeviceDefinitionImageCommandHandlerSuite) TestUpdateDeviceDefinitionImageCommand_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForUpdateImage(s.T(), s.pdb, mk, model, year)

	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionImageCommand{
		DeviceDefinitionID: dd.ID,
		ImageURL:           "https://image.gif",
	})
	result := commandResult.(UpdateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, dd.ID)
}

func (s *UpdateDeviceDefinitionImageCommandHandlerSuite) TestUpdateDeviceDefinitionImageCommand_Exception() {
	ctx := context.Background()

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateDeviceDefinitionImageCommand{
		DeviceDefinitionID: "dd.ID",
	})

	s.Nil(commandResult)
	s.Error(err)
}

func setupDeviceDefinitionForUpdateImage(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}
