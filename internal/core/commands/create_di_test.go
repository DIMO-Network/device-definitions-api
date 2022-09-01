package commands

import (
	"context"
	_ "embed"
	"testing"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
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

	ctrl           *gomock.Controller
	pdb            db.Store
	mockRepository *repositoryMock.MockDeviceIntegrationRepository
	ctx            context.Context

	queryHandler CreateDeviceIntegrationCommandHandler
}

func TestCreateDeviceIntegrationCommandHandler(t *testing.T) {
	suite.Run(t, new(CreateDeviceIntegrationCommandHandlerSuite))
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockRepository = repositoryMock.NewMockDeviceIntegrationRepository(s.ctrl)

	s.queryHandler = NewCreateDeviceIntegrationCommandHandler(s.mockRepository)
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CreateDeviceIntegrationCommandHandlerSuite) TestCreateDeviceIntegrationCommand_Success() {
	ctx := context.Background()

	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	integrationID := "Hummer"
	region := "es-Us"

	di := &models.DeviceIntegration{
		DeviceDefinitionID: deviceDefinitionID,
		IntegrationID:      integrationID,
		Region:             region,
	}

	s.mockRepository.EXPECT().Create(gomock.Any(), deviceDefinitionID, integrationID, region).Return(di, nil).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceIntegrationCommand{
		DeviceDefinitionID: deviceDefinitionID,
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

	s.mockRepository.EXPECT().Create(gomock.Any(), deviceDefinitionID, integrationID, region).Return(nil, errors.New("Error")).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceIntegrationCommand{
		DeviceDefinitionID: deviceDefinitionID,
		IntegrationID:      integrationID,
		Region:             region,
	})

	s.Nil(commandResult)
	s.Error(err)
}
