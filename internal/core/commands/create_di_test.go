package commands

import (
	"context"
	_ "embed"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"

	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"

	"testing"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/testcontainers/testcontainers-go"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateDeviceIntegrationCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                           *gomock.Controller
	pdb                            db.Store
	container                      testcontainers.Container
	mockRepository                 *repositoryMock.MockDeviceIntegrationRepository
	mockDeviceDefinitionRepository *repositoryMock.MockDeviceDefinitionRepository
	mockDeviceDefinitionCache      *mockService.MockDeviceDefinitionCacheService
	ctx                            context.Context

	queryHandler CreateDeviceIntegrationCommandHandler
}

func TestCreateDeviceIntegrationCommandHandler(t *testing.T) {
	suite.Run(t, new(CreateDeviceIntegrationCommandHandlerSuite))
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) SetupTest() {

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.mockDeviceDefinitionRepository = repositoryMock.NewMockDeviceDefinitionRepository(s.ctrl)
	s.mockRepository = repositoryMock.NewMockDeviceIntegrationRepository(s.ctrl)

	s.queryHandler = NewCreateDeviceIntegrationCommandHandler(s.mockRepository, s.pdb.DBS, s.mockDeviceDefinitionCache, s.mockDeviceDefinitionRepository)
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) TestCreateDeviceIntegrationCommand_Success() {
	ctx := context.Background()

	integrationID := "Hummer"
	region := "es-Us"

	// using real DB for integration test
	model := "Testla"
	mk := "Toyota"
	year := 2020
	dd := setupDeviceDefinitionForUpdate(s.T(), s.pdb, mk, model, year)
	repo := repositories.NewDeviceDefinitionRepository(s.pdb.DBS)
	cmdHandler := NewCreateDeviceIntegrationCommandHandler(s.mockRepository, s.pdb.DBS, s.mockDeviceDefinitionCache, repo)

	di := &models.DeviceIntegration{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      integrationID,
		Region:             region,
	}

	s.mockRepository.EXPECT().Create(gomock.Any(), dd.ID, integrationID, region, nil).Return(di, nil).Times(1)

	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByID(ctx, gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	s.mockDeviceDefinitionCache.EXPECT().DeleteDeviceDefinitionCacheBySlug(ctx, gomock.Any(), gomock.Any()).Times(1)

	commandResult, err := cmdHandler.Handle(ctx, &CreateDeviceIntegrationCommand{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      integrationID,
		Region:             region,
	})
	result := commandResult.(CreateDeviceIntegrationCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, integrationID)
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) TestCreateDeviceIntegrationCommand_Exception() {
	ctx := context.Background()

	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	integrationID := "Hummer"
	region := "es-Us"

	s.mockRepository.EXPECT().Create(gomock.Any(), deviceDefinitionID, integrationID, region, nil).Return(nil, errors.New("Error")).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceIntegrationCommand{
		DeviceDefinitionID: deviceDefinitionID,
		IntegrationID:      integrationID,
		Region:             region,
	})

	s.Nil(commandResult)
	s.Error(err)
}
