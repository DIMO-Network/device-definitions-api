package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetDeviceDefinitionByIDsQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	mockDeviceDefinitionCache *mockService.MockDeviceDefinitionCacheService

	queryHandler GetDeviceDefinitionByIdsQueryHandler
}

func TestGetDeviceDefinitionByIdsQueryHandler(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionByIDsQueryHandlerSuite))
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)
	logger := dbtest.Logger()
	s.queryHandler = NewGetDeviceDefinitionByIdsQueryHandler(s.mockDeviceDefinitionCache, logger)
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TestGetDeviceDefinitionByIds_Success() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	mk := "Toyota"
	makeID := "1"
	model := "Hummer"
	year := 2020

	dd := &models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID: deviceDefinitionID,
		DeviceMake: models.DeviceMake{
			ID:   makeID,
			Name: mk,
		},
		Type: models.DeviceType{
			Model: model,
			Year:  year,
			Make:  mk,
		},
		Verified: true,
	}

	s.mockDeviceDefinitionCache.EXPECT().GetDeviceDefinitionByID(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIdsQuery{
		DeviceDefinitionID: []string{deviceDefinitionID},
	})
	result := qryResult.(*grpc.GetDeviceDefinitionResponse)

	s.NoError(err)
	s.Equal(result.DeviceDefinitions[0].DeviceDefinitionId, deviceDefinitionID)
	s.Equal(result.DeviceDefinitions[0].Type.Model, model)
	s.Equal(result.DeviceDefinitions[0].Type.Make, mk)
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TestGetDeviceDefinitionByIds_BadRequest_Exception() {
	ctx := context.Background()

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIdsQuery{
		DeviceDefinitionID: []string{},
	})

	s.Nil(qryResult)
	s.Error(err)

}
