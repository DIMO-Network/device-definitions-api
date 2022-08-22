package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
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
	integrationID := "2D5YSfCcPYW4pTs3NaaqDioUyyl-INT"
	vendor := "AutoPI"
	style := ""
	makeID := "1"
	mk := "Toyota"
	model := "Hummer"

	dd := &models.DeviceDefinition{
		ID:    deviceDefinitionID,
		Model: model,
		Year:  2000,
	}

	di := &models.DeviceIntegration{
		DeviceDefinitionID: deviceDefinitionID,
		IntegrationID:      integrationID,
		Region:             "east-us",
	}
	di.R = di.R.NewStruct()
	di.R.Integration = &models.Integration{ID: "1", Type: "", Style: style, Vendor: vendor}

	dd.R = dd.R.NewStruct()
	dd.R.DeviceIntegrations = models.DeviceIntegrationSlice{di}
	dd.R.DeviceMake = &models.DeviceMake{ID: makeID, Name: mk}

	s.mock_Repository.EXPECT().GetById(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByIdQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})
	result := qryResult.(GetDeviceDefinitionByIDQueryResult)

	s.NoError(err)
	s.Equal(result.DeviceDefinitionID, deviceDefinitionID)
	s.Equal(result.Type.Model, model)
	s.Equal(result.Type.Make, mk)
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
