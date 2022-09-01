package commands

import (
	"context"
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateDeviceDefinitionCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl           *gomock.Controller
	mockRepository *repositoryMock.MockDeviceDefinitionRepository
	ctx            context.Context

	queryHandler CreateDeviceDefinitionCommandHandler
}

func TestCreateDeviceDefinitionCommandHandler(t *testing.T) {
	suite.Run(t, new(CreateDeviceDefinitionCommandHandlerSuite))
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockRepository = repositoryMock.NewMockDeviceDefinitionRepository(s.ctrl)
	s.queryHandler = NewCreateDeviceDefinitionCommandHandler(s.mockRepository)
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TestCreateDeviceDefinitionCommand_Success() {
	ctx := context.Background()

	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	dd := &models.DeviceDefinition{
		ID:    deviceDefinitionID,
		Model: model,
		Year:  int16(year),
	}

	s.mockRepository.EXPECT().GetOrCreate(gomock.Any(), source, mk, model, year).Return(dd, nil).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceDefinitionCommand{
		Source: source,
		Model:  model,
		Make:   mk,
		Year:   year,
	})
	result := commandResult.(CreateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, deviceDefinitionID)
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TestCreateDeviceDefinitionCommand_Exception() {
	ctx := context.Background()

	model := "Hummer"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	s.mockRepository.EXPECT().GetOrCreate(gomock.Any(), source, mk, model, year).Return(nil, errors.New("Error")).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceDefinitionCommand{
		Source: source,
		Model:  model,
		Make:   mk,
		Year:   year,
	})

	s.Nil(commandResult)
	s.Error(err)
}
