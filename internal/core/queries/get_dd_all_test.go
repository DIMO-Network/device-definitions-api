package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetDeviceDefinitionAllQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl           *gomock.Controller
	mockRepository *mocks.MockDeviceDefinitionRepository

	queryHandler GetAllDeviceDefinitionQueryHandler
}

func TestGetDeviceDefinitionAllQueryHandler(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionAllQueryHandlerSuite))
}

func (s *GetDeviceDefinitionAllQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockRepository = mocks.NewMockDeviceDefinitionRepository(s.ctrl)
	s.queryHandler = NewGetAllDeviceDefinitionQueryHandler(s.mockRepository)
}

func (s *GetDeviceDefinitionAllQueryHandlerSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionAllQueryHandlerSuite) TestGetDeviceDefinitionAll_Success() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	mk := "Toyota"
	makeID := "1"
	model := "Hummer"
	year := 2020

	dd := &models.DeviceDefinition{
		ID:       deviceDefinitionID,
		Model:    model,
		Year:     int16(year),
		Verified: true,
	}

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &models.DeviceMake{
		ID:   makeID,
		Name: mk,
	}
	dd.R.DeviceType = &models.DeviceType{
		ID: ksuid.New().String(),
	}
	dd.R.DeviceStyles = append(dd.R.DeviceStyles, &models.DeviceStyle{
		ID:                 ksuid.New().String(),
		ExternalStyleID:    ksuid.New().String(),
		DeviceDefinitionID: deviceDefinitionID,
		Name:               "Hard Top 4dr SUV AWD",
		Source:             "edmunds",
		SubModel:           "Hard Top",
	})

	deviceIntegration := &models.DeviceIntegration{
		DeviceDefinitionID: deviceDefinitionID,
		IntegrationID:      ksuid.New().String(),
		Region:             "Asia",
	}
	deviceIntegration.R = deviceIntegration.R.NewStruct()
	deviceIntegration.R.Integration = &models.Integration{
		ID:     ksuid.New().String(),
		Vendor: "Azure",
		Type:   "test",
		Style:  "test",
	}

	dd.R.DeviceIntegrations = append(dd.R.DeviceIntegrations, deviceIntegration)

	var dds []*models.DeviceDefinition
	dds = append(dds, dd)

	s.mockRepository.EXPECT().GetAll(ctx).Return(dds, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllDeviceDefinitionQuery{})
	result := qryResult.(*grpc.GetDeviceDefinitionResponse)

	s.NoError(err)
	s.Equal(result.DeviceDefinitions[0].DeviceDefinitionId, deviceDefinitionID)

	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].Name, dd.R.DeviceStyles[0].Name)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].ExternalStyleId, dd.R.DeviceStyles[0].ExternalStyleID)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].Source, dd.R.DeviceStyles[0].Source)
	s.Equal(result.DeviceDefinitions[0].DeviceStyles[0].SubModel, dd.R.DeviceStyles[0].SubModel)

}
