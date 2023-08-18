package commands

import (
	"context"
	_ "embed"

	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"

	"testing"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/segmentio/ksuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

//go:embed device_type_vehicle_properties.json
var deviceTypeVehiclePropertyDataSample []byte

type CreateDeviceDefinitionCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	pdb                       db.Store
	container                 testcontainers.Container
	mockRepository            *repositoryMock.MockDeviceDefinitionRepository
	mockPowerTrainTypeService *mock_services.MockPowerTrainTypeService
	ctx                       context.Context

	queryHandler CreateDeviceDefinitionCommandHandler
}

func TestCreateDeviceDefinitionCommandHandler(t *testing.T) {
	suite.Run(t, new(CreateDeviceDefinitionCommandHandlerSuite))
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.mockRepository = repositoryMock.NewMockDeviceDefinitionRepository(s.ctrl)
	s.mockPowerTrainTypeService = mock_services.NewMockPowerTrainTypeService(s.ctrl)
	s.queryHandler = NewCreateDeviceDefinitionCommandHandler(s.mockRepository, s.pdb.DBS, s.mockPowerTrainTypeService)
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TestCreateDeviceDefinitionCommand_Success() {
	ctx := context.Background()

	deviceType := setupCreateDeviceType(s.T(), s.pdb)

	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	deviceMakeID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	model := "Hummer"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	dd := &models.DeviceDefinition{
		ID:           deviceDefinitionID,
		Model:        model,
		Year:         int16(year),
		DeviceTypeID: null.StringFrom(deviceType.ID),
	}
	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &models.DeviceMake{ID: deviceMakeID, Name: mk}

	iceValue := "ICE"
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), nil, null.JSON{}, null.JSON{}).Return(iceValue, nil)
	s.mockRepository.EXPECT().GetOrCreate(gomock.Any(), nil, source, "", mk, model, year, gomock.Any(), gomock.Any(), false, gomock.Any()).Return(dd, nil).Times(1)

	var deviceAttributes []*coremodels.UpdateDeviceTypeAttribute

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceDefinitionCommand{
		Source:       source,
		Model:        model,
		Make:         mk,
		Year:         year,
		DeviceTypeID: deviceType.ID,
		DeviceAttributes: append(deviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "MPG",
			Value: "12",
		}),
	})
	result := commandResult.(CreateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, deviceDefinitionID)
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TestCreateDeviceDefinitionCommand_Exception() {
	ctx := context.Background()

	deviceType := setupCreateDeviceType(s.T(), s.pdb)

	model := "Hummer"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	iceValue := "ICE"
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), nil, null.JSON{}, null.JSON{}).Return(iceValue, nil)
	s.mockRepository.EXPECT().
		GetOrCreate(gomock.Any(), nil, source, "", mk, model, year, gomock.Any(), gomock.Any(), false, gomock.Any()).Return(nil, errors.New("Error")).Times(1)

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceDefinitionCommand{
		Source:       source,
		Model:        model,
		Make:         mk,
		Year:         year,
		DeviceTypeID: deviceType.ID,
	})

	s.Nil(commandResult)
	s.Error(err)
}

func setupCreateDeviceType(t *testing.T, pdb db.Store) models.DeviceType {
	deviceType := models.DeviceType{
		ID:          ksuid.New().String(),
		Name:        "vehicle",
		Metadatakey: "vehicle_info",
		Properties:  null.JSONFrom(deviceTypeVehiclePropertyDataSample),
	}
	err := deviceType.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return deviceType
}
