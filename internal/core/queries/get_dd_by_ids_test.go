package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetDeviceDefinitionByIDsQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	mockDeviceDefinitionCache *mockService.MockDeviceDefinitionCacheService

	queryHandler GetDeviceDefinitionByIDsQueryHandler
}

func TestGetDeviceDefinitionByIdsQueryHandler(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionByIDsQueryHandlerSuite))
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)
	logger := dbtest.Logger()
	s.queryHandler = NewGetDeviceDefinitionByIDsQueryHandler(s.mockDeviceDefinitionCache, logger)
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TestGetDeviceDefinitionByIds_Success() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	mk := "Toyota"
	makeID := "1"

	dd := &models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID: deviceDefinitionID,
		DeviceMake: models.DeviceMake{
			ID:   makeID,
			Name: mk,
		},
		DeviceStyles: []models.DeviceStyle{
			models.DeviceStyle{
				ID:                 ksuid.New().String(),
				ExternalStyleID:    ksuid.New().String(),
				DeviceDefinitionID: deviceDefinitionID,
				Name:               "Hard Top 4dr SUV AWD",
				Source:             "edmunds",
				SubModel:           "Hard Top",
			},
			models.DeviceStyle{
				ID:                 ksuid.New().String(),
				ExternalStyleID:    ksuid.New().String(),
				DeviceDefinitionID: deviceDefinitionID,
				Name:               "4dr SUV 4WD",
				Source:             "edmunds",
				SubModel:           "Wagon",
			},
		},
		Verified: true,
	}

	s.mockDeviceDefinitionCache.EXPECT().GetDeviceDefinitionByID(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIDsQuery{
		DeviceDefinitionID: []string{deviceDefinitionID},
	})
	result := qryResult.(*grpc.GetDeviceDefinitionResponse)

	s.NoError(err)
	s.Equal(result.DeviceDefinitions[0].DeviceDefinitionId, deviceDefinitionID)

	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].DeviceDefinitionId, dd.DeviceDefinitionID)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].Name, dd.DeviceStyles[0].Name)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].ExternalStyleId, dd.DeviceStyles[0].ExternalStyleID)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].Source, dd.DeviceStyles[0].Source)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].SubModel, dd.DeviceStyles[0].SubModel)
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TestGetDeviceDefinitionByIds_BadRequest_Exception() {
	ctx := context.Background()

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIDsQuery{
		DeviceDefinitionID: []string{},
	})

	s.Nil(qryResult)
	s.Error(err)

}
