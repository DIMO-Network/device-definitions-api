package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetAllDeviceDefinitionQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                *gomock.Controller
	mock_Repository     *mocks.MockDeviceDefinitionRepository
	mock_MakeRepository *mocks.MockDeviceMakeRepository

	queryHandler GetAllDeviceDefinitionQueryHandler
}

func TestGetAllDeviceDefinitionQueryHandler(t *testing.T) {
	suite.Run(t, new(GetAllDeviceDefinitionQueryHandlerSuite))
}

func (s *GetAllDeviceDefinitionQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mock_Repository = mocks.NewMockDeviceDefinitionRepository(s.ctrl)
	s.mock_MakeRepository = mocks.NewMockDeviceMakeRepository(s.ctrl)

	s.queryHandler = NewGetAllDeviceDefinitionQueryHandler(s.mock_Repository, s.mock_MakeRepository)
}

func (s *GetAllDeviceDefinitionQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetAllDeviceDefinitionQueryHandlerSuite) TestGetAll_Success() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	makeId := "1"
	make := "Toyota"

	var dd []models.DeviceDefinition
	dd = append(dd, models.DeviceDefinition{
		ID:           deviceDefinitionID,
		Model:        model,
		Year:         2000,
		DeviceMakeID: makeId,
	})

	var makes []models.DeviceMake
	makes = append(makes, models.DeviceMake{
		ID:   makeId,
		Name: make,
	})

	s.mock_MakeRepository.EXPECT().GetAll(ctx).Return(makes, nil).Times(1)
	s.mock_Repository.EXPECT().GetAll(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllDeviceDefinitionQuery{})
	result := qryResult.(*GetAllDeviceDefinitionQueryResult)

	s.NoError(err)
	s.Equal(result.Make, make)
}

func (s *GetAllDeviceDefinitionQueryHandlerSuite) TestGetAll_Fail() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	makeId := "1"
	make := "Toyota"

	var dd []models.DeviceDefinition
	dd = append(dd, models.DeviceDefinition{
		ID:           deviceDefinitionID,
		Model:        model,
		Year:         2000,
		DeviceMakeID: makeId,
	})

	var makes []models.DeviceMake
	makes = append(makes, models.DeviceMake{
		ID:   makeId,
		Name: make,
	})

	s.mock_Repository.EXPECT().GetAll(ctx, true).Return(dd, nil).Times(1)
	s.mock_MakeRepository.EXPECT().GetAll(ctx).Return(makes, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllDeviceDefinitionQuery{})

	s.Nil(qryResult)
	s.Error(err)
}
