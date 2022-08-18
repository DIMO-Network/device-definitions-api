package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GetDeviceDefinitionByMakeModelYearQuerySuite struct {
	suite.Suite
	*require.Assertions

	ctrl            *gomock.Controller
	mock_Repository *mocks.MockDeviceDefinitionRepository

	queryHandler GetDeviceDefinitionByMakeModelYearQueryHandler
}

func TestGetDeviceDefinitionByMakeModelYearQuery(t *testing.T) {
	suite.Run(t, new(GetDeviceDefinitionByMakeModelYearQuerySuite))
}

func (s *GetDeviceDefinitionByMakeModelYearQuerySuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.mock_Repository = mocks.NewMockDeviceDefinitionRepository(s.ctrl)

	s.queryHandler = NewGetDeviceDefinitionByMakeModelYearQueryHandler(s.mock_Repository)
}

func (s *GetDeviceDefinitionByMakeModelYearQuerySuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *GetDeviceDefinitionByMakeModelYearQuerySuite) TestGetDeviceDefinitionByMakeModelYear_With_Items() {
	ctx := context.Background()
	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	makeId := "1"
	mk := "Toyota"
	year := 2020

	dd := &models.DeviceDefinition{
		ID:           deviceDefinitionID,
		Model:        model,
		Year:         int16(year),
		DeviceMakeID: makeId,
		Verified:     true,
	}
	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &models.DeviceMake{ID: makeId, Name: mk}

	s.mock_Repository.EXPECT().GetByMakeModelAndYears(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &GetDeviceDefinitionByMakeModelYearQuery{
		Make:  mk,
		Model: model,
		Year:  year,
	})
	result := qryResult.(GetDeviceDefinitionByMakeModelYearQueryResult)

	s.NoError(err)
	assert.Equal(s.T(), deviceDefinitionID, result.DeviceDefinitionID)
	assert.Equal(s.T(), mk, result.Type.Make)
	assert.Equal(s.T(), model, result.Type.Model)
	assert.Equal(s.T(), year, result.Type.Year)
}
