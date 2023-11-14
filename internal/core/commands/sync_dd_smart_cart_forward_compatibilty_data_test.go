package commands

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	gatewayMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type SyncSmartCartForwardCompatibilityCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                *gomock.Controller
	pdb                 db.Store
	container           testcontainers.Container
	mockSmartCarService *gatewayMock.MockSmartCarService
	mockRepository      *repositoryMock.MockDeviceDefinitionRepository
	ctx                 context.Context

	queryHandler SyncSmartCartForwardCompatibilityCommandHandler
}

func TestSyncSmartCartForwardCompatibilityCommandHandler(t *testing.T) {
	suite.Run(t, new(SyncSmartCartForwardCompatibilityCommandHandlerSuite))
}

func (s *SyncSmartCartForwardCompatibilityCommandHandlerSuite) SetupTest() {

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

	s.queryHandler = NewSyncSmartCartForwardCompatibilityCommandHandler(s.pdb.DBS, s.mockSmartCarService, s.mockRepository)
}

func (s *SyncSmartCartForwardCompatibilityCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *SyncSmartCartForwardCompatibilityCommandHandlerSuite) TestSyncSmartCartForwardCompatibilityCommand() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020

	dd := setupDeviceDefinitionForSmartCarForwardCompatibility(s.T(), s.pdb, mk, model, year)
	integration := setupIntegrationForSmartCarForwardCompatibility(s.T(), s.pdb)
	deviceIntegration := setupDeviceDefinitionIntegrationForSmartCarForwardCompatibility(s.T(), s.pdb, dd, integration)

	dd.R = dd.R.NewStruct()
	dd.R.DeviceIntegrations = models.DeviceIntegrationSlice{deviceIntegration}
	s.mockSmartCarService.EXPECT().GetOrCreateSmartCarIntegration(gomock.Any()).Return(integration.ID, nil).Times(1)
	s.mockRepository.EXPECT().GetByMakeModelAndYears(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any()).Return(dd, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncSearchDataCommand{})
	require.NoError(s.T(), err)
	require.NotNilf(s.T(), qryResult, "query result cannot be nil")
	result := qryResult.(SyncSmartCartForwardCompatibilityCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.Status, true)
}

func setupDeviceDefinitionForSmartCarForwardCompatibility(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}

func setupIntegrationForSmartCarForwardCompatibility(t *testing.T, pdb db.Store) *models.Integration {
	i := dbtesthelper.SetupCreateSmartCarIntegration(t, pdb)
	return i
}

func setupDeviceDefinitionIntegrationForSmartCarForwardCompatibility(t *testing.T, pdb db.Store, dd *models.DeviceDefinition, i *models.Integration) *models.DeviceIntegration {
	di := dbtesthelper.SetupCreateDeviceIntegration(t, dd, i.ID, "Americas", pdb)
	return di
}
