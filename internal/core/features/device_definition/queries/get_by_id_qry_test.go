package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories/mocks"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetByIdQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl            *gomock.Controller
	mock_Repository *mocks.MockIDeviceDefinitionRepository

	queryHandler GetByIdQueryHandler
}

func TestAddWithDeviceIdCommandHandlerSuite(t *testing.T) {
	suite.Run(t, new(GetByIdQueryHandlerSuite))
}

func (s *GetByIdQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mock_Repository = mocks.NewMockIDeviceDefinitionRepository(s.ctrl)

	s.queryHandler = NewGetByIdQueryHandler(s.mock_Repository)
}

func (s *GetByIdQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetByIdQueryHandlerSuite) TestGetById_Success() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	dd := &models.DeviceDefinition{
		ID:    deviceDefinitionID,
		Model: "Hummer",
		Year:  2000,
	}

	s.mock_Repository.EXPECT().GetById(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetByIdQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})
	result := qryResult.(*GetByIdQueryResult)

	s.NoError(err)
	s.Equal(result.DeviceDefinitionID, dd.ID)
	s.Equal(result.Model, dd.Model)
	s.Equal(result.Year, dd.Year)
}

func (s *GetByIdQueryHandlerSuite) TestGetById_Fail() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	s.mock_Repository.EXPECT().GetById(ctx, gomock.Any()).Return(nil, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetByIdQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})

	s.Nil(qryResult)
	s.Error(err)
	s.EqualError(err, fmt.Sprintf("could not find device definition id: %s", deviceDefinitionID))
}
