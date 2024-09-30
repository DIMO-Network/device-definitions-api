package commands

import (
	"context"
	_ "embed"

	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"

	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"

	"testing"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/testcontainers/testcontainers-go"

	"github.com/pkg/errors"

	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CreateDeviceIntegrationCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                               *gomock.Controller
	pdb                                db.Store
	container                          testcontainers.Container
	mockRepository                     *repositoryMock.MockDeviceIntegrationRepository
	mockDeviceDefinitionRepository     *repositoryMock.MockDeviceDefinitionRepository
	mockDeviceDefinitionCache          *mockService.MockDeviceDefinitionCacheService
	ctx                                context.Context
	mockDeviceDefinitionOnChainService *mock_gateways.MockDeviceDefinitionOnChainService

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
	s.mockDeviceDefinitionOnChainService = mock_gateways.NewMockDeviceDefinitionOnChainService(s.ctrl)

	s.queryHandler = NewCreateDeviceIntegrationCommandHandler(s.mockRepository, s.pdb.DBS, s.mockDeviceDefinitionCache, s.mockDeviceDefinitionRepository)
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
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
