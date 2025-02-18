package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mockService "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetDeviceDefinitionByMakeModelYearQuerySuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	mockDeviceDefinitionCache *mockService.MockDeviceDefinitionCacheService

	queryHandler GetDeviceDefinitionByMakeModelYearQueryHandler
}

func TestGetDeviceDefinitionByMakeModelYearQuery(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionByMakeModelYearQuerySuite))
}

func (s *GetDeviceDefinitionByMakeModelYearQuerySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mockDeviceDefinitionCache = mockService.NewMockDeviceDefinitionCacheService(s.ctrl)

	s.queryHandler = NewGetDeviceDefinitionByMakeModelYearQueryHandler(s.mockDeviceDefinitionCache)
}

func (s *GetDeviceDefinitionByMakeModelYearQuerySuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionByMakeModelYearQuerySuite) TestGetDeviceDefinitionByMakeModelYear_With_Items() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	makeID := "1"
	mk := "Toyota"
	year := 2020

	dd := &models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID: deviceDefinitionID,
		DeviceMake: models.DeviceMake{
			ID:   makeID,
			Name: mk,
		},
		Verified: true,
	}

	s.mockDeviceDefinitionCache.EXPECT().GetDeviceDefinitionByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByMakeModelYearQuery{
		Make:  mk,
		Model: model,
		Year:  year,
	})
	result := qryResult.(*models.GetDeviceDefinitionQueryResult)

	s.NoError(err)
	assert.Equal(s.T(), deviceDefinitionID, result.DeviceDefinitionID)
}
