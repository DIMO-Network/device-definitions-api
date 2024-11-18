package commands

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"

	"go.uber.org/mock/gomock"

	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"

	"testing"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

//go:embed device_type_vehicle_properties.json
var deviceTypeVehiclePropertyDataSample []byte

const (
	dbName               = "device_definitions_api"
	migrationsDirRelPath = "../../infrastructure/db/migrations"
)

type CreateDeviceDefinitionCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	pdb                       db.Store
	container                 testcontainers.Container
	mockDeviceDefRepo         *repositoryMock.MockDeviceDefinitionRepository
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

	s.mockDeviceDefRepo = repositoryMock.NewMockDeviceDefinitionRepository(s.ctrl)
	s.mockPowerTrainTypeService = mock_services.NewMockPowerTrainTypeService(s.ctrl)
	s.queryHandler = NewCreateDeviceDefinitionCommandHandler(s.mockDeviceDefRepo, s.pdb.DBS, s.mockPowerTrainTypeService)
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TestCreateDeviceDefinitionCommand_Success() {
	ctx := context.Background()

	deviceType := setupCreateDeviceType(s.T(), s.pdb, common.DefaultDeviceType, "Vehicle Information", "vehicle_info")

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
	s.mockDeviceDefRepo.EXPECT().GetOrCreate(gomock.Any(), nil, source, "", mk, model, year, gomock.Any(), gomock.Any(), false, gomock.Any()).Return(dd, nil).Times(1)

	var deviceAttributes []*coremodels.UpdateDeviceTypeAttribute

	commandResult, err := s.queryHandler.Handle(ctx, &CreateDeviceDefinitionCommand{
		Source:       source,
		Model:        model,
		Make:         mk,
		Year:         year,
		DeviceTypeID: deviceType.ID,
		DeviceAttributes: append(deviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "mpg",
			Value: "12",
		}),
	})
	result := commandResult.(CreateDeviceDefinitionCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, deviceDefinitionID)
}

func (s *CreateDeviceDefinitionCommandHandlerSuite) TestCreateDeviceDefinitionCommand_Exception() {
	ctx := context.Background()

	deviceType := setupCreateDeviceType(s.T(), s.pdb, common.DefaultDeviceType, "Vehicle Information", "vehicle_info")

	model := "Hummer"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	iceValue := "ICE"
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), nil, null.JSON{}, null.JSON{}).Return(iceValue, nil)
	s.mockDeviceDefRepo.EXPECT().
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

func setupCreateDeviceType(t *testing.T, pdb db.Store, id, name, mdKey string) models.DeviceType {
	// Upsert won't update the properties field, delete and re-create, start clean
	dt, _ := models.FindDeviceType(context.Background(), pdb.DBS().Writer, id)
	if dt != nil {
		_, _ = dt.Delete(context.Background(), pdb.DBS().Writer)
	}
	// create dt with properties from a file
	deviceType := models.DeviceType{
		ID:          id,
		Name:        name,
		Metadatakey: mdKey,
		Properties:  null.JSONFrom(deviceTypeVehiclePropertyDataSample),
	}
	err := deviceType.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return deviceType
}
