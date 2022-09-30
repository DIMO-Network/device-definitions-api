package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
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
		DeviceStyles: []models.GetDeviceDefinitionStyles{
			models.GetDeviceDefinitionStyles{
				ID:                 ksuid.New().String(),
				ExternalStyleID:    ksuid.New().String(),
				DeviceDefinitionID: deviceDefinitionID,
				Name:               "Hard Top 4dr SUV AWD",
				Source:             "edmunds",
				SubModel:           "Hard Top",
			},
			models.GetDeviceDefinitionStyles{
				ID:                 ksuid.New().String(),
				ExternalStyleID:    ksuid.New().String(),
				DeviceDefinitionID: deviceDefinitionID,
				Name:               "4dr SUV 4WD",
				Source:             "edmunds",
				SubModel:           "Wagon",
			},
		},
		DeviceIntegrations: []models.GetDeviceDefinitionIntegration{
			models.GetDeviceDefinitionIntegration{
				ID:     ksuid.New().String(),
				Type:   "API",
				Style:  "Webhook",
				Vendor: "SmartCar",
				Region: "Asia",
			},
			models.GetDeviceDefinitionIntegration{
				ID:     ksuid.New().String(),
				Type:   "API",
				Style:  "Webhook",
				Vendor: "SmartCar",
				Region: "USA",
			},
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

	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].DeviceDefinitionId, dd.DeviceDefinitionID)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].Name, dd.DeviceStyles[0].Name)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].ExternalStyleId, dd.DeviceStyles[0].ExternalStyleID)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].Source, dd.DeviceStyles[0].Source)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].SubModel, dd.DeviceStyles[0].SubModel)

	s.Equal(result.DeviceDefinitions[0].DeviceIntegrations[0].Id, dd.DeviceIntegrations[0].ID)
	s.Equal(result.DeviceDefinitions[0].DeviceIntegrations[0].Vendor, dd.DeviceIntegrations[0].Vendor)
	s.Equal(result.DeviceDefinitions[0].DeviceIntegrations[0].Style, dd.DeviceIntegrations[0].Style)
	s.Equal(result.DeviceDefinitions[0].DeviceIntegrations[0].Region, dd.DeviceIntegrations[0].Region)
	s.Equal(result.DeviceDefinitions[0].DeviceIntegrations[0].Country, dd.DeviceIntegrations[0].Country)
}

func (s *GetDeviceDefinitionByIDsQueryHandlerSuite) TestGetDeviceDefinitionByIds_BadRequest_Exception() {
	ctx := context.Background()

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIdsQuery{
		DeviceDefinitionID: []string{},
	})

	s.Nil(qryResult)
	s.Error(err)

}
