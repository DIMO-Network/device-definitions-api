package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetDeviceDefinitionByIDQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	mockDeviceDefinitionCache *mockService.MockDeviceDefinitionCacheService

	queryHandler GetDeviceDefinitionByIDQueryHandler
}

func TestGetDeviceDefinitionByIdQueryHandler(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionByIDQueryHandlerSuite))
}

func (s *GetDeviceDefinitionByIDQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)

	s.queryHandler = NewGetDeviceDefinitionByIDQueryHandler(s.mockDeviceDefinitionCache)
}

func (s *GetDeviceDefinitionByIDQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionByIDQueryHandlerSuite) TestGetDeviceDefinitionById_Success() {
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
		DeviceIntegrations: []models.DeviceIntegration{
			models.DeviceIntegration{
				ID:     ksuid.New().String(),
				Type:   "API",
				Style:  "Webhook",
				Vendor: "SmartCar",
				Region: "Asia",
			},
			models.DeviceIntegration{
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

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIDQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})
	result := qryResult.(*models.GetDeviceDefinitionQueryResult)

	s.NoError(err)
	s.Equal(result.DeviceDefinitionID, deviceDefinitionID)
	s.Equal(result.Type.Model, model)
	s.Equal(result.Type.Make, mk)

	s.Equal(result.DeviceStyles[0].DeviceDefinitionID, dd.DeviceDefinitionID)
	s.Equal(result.DeviceStyles[0].Name, dd.DeviceStyles[0].Name)
	s.Equal(result.DeviceStyles[0].ExternalStyleID, dd.DeviceStyles[0].ExternalStyleID)
	s.Equal(result.DeviceStyles[0].Source, dd.DeviceStyles[0].Source)
	s.Equal(result.DeviceStyles[0].SubModel, dd.DeviceStyles[0].SubModel)

	s.Equal(result.DeviceIntegrations[0].ID, dd.DeviceIntegrations[0].ID)
	s.Equal(result.DeviceIntegrations[0].Vendor, dd.DeviceIntegrations[0].Vendor)
	s.Equal(result.DeviceIntegrations[0].Style, dd.DeviceIntegrations[0].Style)
	s.Equal(result.DeviceIntegrations[0].Region, dd.DeviceIntegrations[0].Region)

}

func (s *GetDeviceDefinitionByIDQueryHandlerSuite) TestGetDeviceDefinitionById_Exception() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	s.mockDeviceDefinitionCache.EXPECT().GetDeviceDefinitionByID(ctx, gomock.Any()).Return(nil, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIDQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})

	s.Nil(qryResult)
	s.Error(err)
}
