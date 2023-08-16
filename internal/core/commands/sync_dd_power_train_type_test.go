package commands

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"os"
	"testing"

	mock_service "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type SyncPowerTrainTypeCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	pdb                       db.Store
	container                 testcontainers.Container
	mockPowerTrainTypeService *mock_service.MockPowerTrainTypeService
	ctx                       context.Context

	queryHandler SyncPowerTrainTypeCommandHandler
}

func TestSyncPowerTrainTypeCommandHandler(t *testing.T) {
	suite.Run(t, new(SyncPowerTrainTypeCommandHandlerSuite))
}

func (s *SyncPowerTrainTypeCommandHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockPowerTrainTypeService = mock_service.NewMockPowerTrainTypeService(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewSyncPowerTrainTypeCommandHandler(s.pdb.DBS, zerolog.New(os.Stdout), s.mockPowerTrainTypeService)
}

func (s *SyncPowerTrainTypeCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *SyncPowerTrainTypeCommandHandlerSuite) TestSyncPowerTrainTypeCommand_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)

	ICE := "ICE"
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&ICE, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncPowerTrainTypeCommand{DeviceTypeID: dd.DeviceTypeID.String})
	require.NoError(s.T(), err, "handler failed to execute")

	result := qryResult.(SyncPowerTrainTypeCommandResult)

	assert.Equal(s.T(), result.Status, true)

	ddUpdated, _ := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.EQ(dd.ID),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).One(ctx, s.pdb.DBS().Writer)

	metadataKey := ddUpdated.R.DeviceType.Metadatakey
	var metadataAttributes map[string]any
	if err = ddUpdated.Metadata.Unmarshal(&metadataAttributes); err == nil {
		for key, value := range metadataAttributes[metadataKey].(map[string]interface{}) {
			if key == common.PowerTrainType {
				assert.Equal(s.T(), value, ICE)
				break
			}
		}

	}

}

func (s *SyncPowerTrainTypeCommandHandlerSuite) TestSyncPowerTrainTypeCommand_With_Force_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForPowerTrain(s.T(), s.pdb, mk, model, year)

	HEV := "HEV"
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&HEV, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncPowerTrainTypeCommand{ForceUpdate: true, DeviceTypeID: dd.DeviceTypeID.String})
	require.NoError(s.T(), err, "handler failed to execute")

	result := qryResult.(SyncPowerTrainTypeCommandResult)

	assert.Equal(s.T(), result.Status, true)

	ddUpdated, _ := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.EQ(dd.ID),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).One(ctx, s.pdb.DBS().Writer)

	metadataKey := ddUpdated.R.DeviceType.Metadatakey
	var metadataAttributes map[string]any
	if err = ddUpdated.Metadata.Unmarshal(&metadataAttributes); err == nil {
		for key, value := range metadataAttributes[metadataKey].(map[string]interface{}) {
			if key == common.PowerTrainType {
				assert.Equal(s.T(), value, HEV)
				break
			}
		}

	}

}

func setupDeviceDefinitionForPowerTrain(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfoIncludePowerTrain(t, dm, modelName, year, pdb)
	img := models.Image{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: dd.ID,
		Width:              null.IntFrom(640),
		Height:             null.IntFrom(480),
		SourceURL:          "https://some-image.com/img.jpg",
	}
	err := img.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	return dd
}

func setupDeviceDefinition(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(t, dm, modelName, year, pdb)
	img := models.Image{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: dd.ID,
		Width:              null.IntFrom(640),
		Height:             null.IntFrom(480),
		SourceURL:          "https://some-image.com/img.jpg",
	}
	err := img.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	return dd
}
