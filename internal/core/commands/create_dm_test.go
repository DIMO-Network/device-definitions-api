package commands

import (
	"context"
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CreateDeviceMakeCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl           *gomock.Controller
	mockRepository *repositoryMock.MockDeviceMakeRepository
	ctx            context.Context

	queryHandler CreateDeviceMakeCommandHandler
}

func TestCreateDeviceMakeCommandHandler(t *testing.T) {
	suite.Run(t, new(CreateDeviceMakeCommandHandlerSuite))
}

func (s *CreateDeviceMakeCommandHandlerSuite) SetupTest() {

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockRepository = repositoryMock.NewMockDeviceMakeRepository(s.ctrl)

	s.queryHandler = NewCreateDeviceMakeCommandHandler(s.mockRepository)
}

func (s *CreateDeviceMakeCommandHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CreateDeviceMakeCommandHandlerSuite) TestCreateDeviceMakeCommand_Success() {
	ctx := context.Background()

	name := "Ford"
	templateID := "01"

	dm := &models.DeviceMake{
		ID:   ksuid.New().String(),
		Name: name,
	}

	s.mockRepository.EXPECT().GetOrCreate(gomock.Any(), name, gomock.Any(), gomock.Any(), gomock.Any(), templateID).Return(dm, nil).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceMakeCommand{
		Name:               name,
		HardwareTemplateID: templateID,
	})
	result := commandResult.(CreateDeviceMakeCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, dm.ID)
}
