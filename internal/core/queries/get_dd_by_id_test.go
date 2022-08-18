package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetDeviceDefinitionByIdQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl            *gomock.Controller
	mock_Repository *mocks.MockDeviceDefinitionRepository

	queryHandler GetDeviceDefinitionByIdQueryHandler
}

func TestGetDeviceDefinitionByIdQueryHandler(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionByIdQueryHandlerSuite))
}

func (s *GetDeviceDefinitionByIdQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mock_Repository = mocks.NewMockDeviceDefinitionRepository(s.ctrl)

	s.queryHandler = NewGetDeviceDefinitionByIdQueryHandler(s.mock_Repository)
}

func (s *GetDeviceDefinitionByIdQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionByIdQueryHandlerSuite) TestGetDeviceDefinitionById_Success() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	dd := &models.DeviceDefinition{
		ID:    deviceDefinitionID,
		Model: "Hummer",
		Year:  2000,
	}

	s.mock_Repository.EXPECT().GetById(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIdQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})
	result := qryResult.(*GetDeviceDefinitionByIdQueryResult)

	s.NoError(err)
	s.Equal(result.DeviceDefinitionID, dd.ID)
	s.Equal(result.Model, dd.Model)
	s.Equal(result.Year, dd.Year)
}

func (s *GetDeviceDefinitionByIdQueryHandlerSuite) TestGetDeviceDefinitionById_Exception() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	s.mock_Repository.EXPECT().GetById(ctx, gomock.Any()).Return(nil, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIdQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})

	s.Nil(qryResult)
	s.Error(err)
	s.EqualError(err, fmt.Sprintf("could not find device definition id: %s", deviceDefinitionID))
}
