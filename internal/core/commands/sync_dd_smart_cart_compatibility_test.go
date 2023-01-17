package commands

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	gatewayMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

//go:embed test_smart_data.json
var smartDataSample []byte

type SyncSmartCartCompatibilityCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                *gomock.Controller
	pdb                 db.Store
	container           testcontainers.Container
	mockSmartCarService *gatewayMock.MockSmartCarService
	mockRepository      *repositoryMock.MockDeviceDefinitionRepository
	ctx                 context.Context

	queryHandler SyncSmartCartCompatibilityCommandHandler
}

func TestSyncSmartCartCompatibilityCommandHandler(t *testing.T) {
	suite.Run(t, new(SyncSmartCartCompatibilityCommandHandlerSuite))
}

func (s *SyncSmartCartCompatibilityCommandHandlerSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockSmartCarService = gatewayMock.NewMockSmartCarService(s.ctrl)
	s.mockRepository = repositoryMock.NewMockDeviceDefinitionRepository(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewSyncSmartCartCompatibilityCommandHandler(s.pdb.DBS, s.mockSmartCarService, s.mockRepository)
}

func (s *SyncSmartCartCompatibilityCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *SyncSmartCartCompatibilityCommandHandlerSuite) TestSyncSmartCartCompatibilityCommand() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForSmartCarCompatibility(s.T(), s.pdb, mk, model, year)
	integration := setupIntegrationForSmartCarCompatibility(s.T(), s.pdb)
	deviceIntegration := setupDeviceDefinitionIntegrationForSmartCarCompatibility(s.T(), s.pdb, dd, integration)

	dd.R = dd.R.NewStruct()
	dd.R.DeviceIntegrations = models.DeviceIntegrationSlice{deviceIntegration}

	smartCarCompatibilityData := &gateways.SmartCarCompatibilityData{}

	_ = json.Unmarshal(smartDataSample, smartCarCompatibilityData)

	s.mockSmartCarService.EXPECT().GetSmartCarVehicleData().Return(smartCarCompatibilityData, nil).Times(1)
	s.mockSmartCarService.EXPECT().GetOrCreateSmartCarIntegration(gomock.Any()).Return(integration.ID, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncSearchDataCommand{})
	result := qryResult.(SyncSmartCartCompatibilityCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.Status, true)
}

func setupDeviceDefinitionForSmartCarCompatibility(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}

func setupIntegrationForSmartCarCompatibility(t *testing.T, pdb db.Store) *models.Integration {
	i := dbtesthelper.SetupCreateSmartCarIntegration(t, pdb)
	return i
}

func setupDeviceDefinitionIntegrationForSmartCarCompatibility(t *testing.T, pdb db.Store, dd *models.DeviceDefinition, i *models.Integration) *models.DeviceIntegration {
	di := dbtesthelper.SetupCreateDeviceIntegration(t, dd, i.ID, "Americas", pdb)
	return di
}
