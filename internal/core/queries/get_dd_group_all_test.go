package queries

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetAllDeviceDefinitionGroupGroupQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl               *gomock.Controller
	mockRepository     *mocks.MockDeviceDefinitionRepository
	mockMakeRepository *mocks.MockDeviceMakeRepository

	queryHandler GetAllDeviceDefinitionGroupQueryHandler
}

func TestGetAllDeviceDefinitionGroupQueryHandler(t *testing.T) {
	suite.Run(t, new(GetAllDeviceDefinitionGroupGroupQueryHandlerSuite))
}

func (s *GetAllDeviceDefinitionGroupGroupQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockRepository = mocks.NewMockDeviceDefinitionRepository(s.ctrl)
	s.mockMakeRepository = mocks.NewMockDeviceMakeRepository(s.ctrl)

	s.queryHandler = NewGetAllDeviceDefinitionGroupQueryHandler(s.mockRepository, s.mockMakeRepository)
}

func (s *GetAllDeviceDefinitionGroupGroupQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetAllDeviceDefinitionGroupGroupQueryHandlerSuite) TestGetAllDeviceDefinitionGroupQuery_With_Items() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	makeID := "1"
	mk := "Toyota"

	var dd []*models.DeviceDefinition
	dd = append(dd, &models.DeviceDefinition{
		ID:           deviceDefinitionID,
		Model:        model,
		Year:         2000,
		DeviceMakeID: makeID,
	})

	var makes []*models.DeviceMake
	makes = append(makes, &models.DeviceMake{
		ID:   makeID,
		Name: mk,
	})

	s.mockMakeRepository.EXPECT().GetAll(ctx).Return(makes, nil).Times(1)
	s.mockRepository.EXPECT().GetAll(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllDeviceDefinitionGroupQuery{})
	result := qryResult.([]GetAllDeviceDefinitionGroupQueryResult)

	s.NoError(err)
	s.Len(result, 1)
	assert.Equal(s.T(), mk, result[0].Make)
}

func (s *GetAllDeviceDefinitionGroupGroupQueryHandlerSuite) TestGetAllDeviceDefinitionGroupQuery_Empty() {
	ctx := context.Background()

	var dd []*models.DeviceDefinition
	var makes []*models.DeviceMake

	s.mockMakeRepository.EXPECT().GetAll(ctx).Return(makes, nil).Times(1)
	s.mockRepository.EXPECT().GetAll(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllDeviceDefinitionGroupQuery{})
	result := qryResult.([]GetAllDeviceDefinitionGroupQueryResult)

	s.NoError(err)
	s.Len(result, 0)
}
