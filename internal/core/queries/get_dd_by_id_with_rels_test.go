package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetDeviceDefinitionWithRelsQuerySuite struct {
	suite.Suite
	*require.Assertions

	ctrl           *gomock.Controller
	mockRepository *mocks.MockDeviceDefinitionRepository

	queryHandler GetDeviceDefinitionWithRelsQueryHandler
}

func TestGetDeviceDefinitionWithRelsQuery(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionWithRelsQuerySuite))
}

func (s *GetDeviceDefinitionWithRelsQuerySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockRepository = mocks.NewMockDeviceDefinitionRepository(s.ctrl)

	s.queryHandler = NewGetDeviceDefinitionWithRelsQueryHandler(s.mockRepository)
}

func (s *GetDeviceDefinitionWithRelsQuerySuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionWithRelsQuerySuite) TestGetDeviceDefinitionWithRels_Empty() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	dd := &models.DeviceDefinition{
		ID:    deviceDefinitionID,
		Model: "Hummer",
		Year:  2000,
	}

	s.mockRepository.EXPECT().GetWithIntegrations(ctx, gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionWithRelsQuery{DeviceDefinitionID: deviceDefinitionID})
	result := qryResult.([]GetDeviceDefinitionWithRelsQueryResult)

	s.NoError(err)
	s.Len(result, 0)
}

func (s *GetDeviceDefinitionWithRelsQuerySuite) TestGetDeviceDefinitionWithRels_Exception() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	s.mockRepository.EXPECT().GetWithIntegrations(ctx, gomock.Any()).Return(nil, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionWithRelsQuery{
		DeviceDefinitionID: deviceDefinitionID,
	})

	s.Nil(qryResult)
	s.Error(err)
	s.EqualError(err, fmt.Sprintf("could not find device definition id: %s", deviceDefinitionID))
}
