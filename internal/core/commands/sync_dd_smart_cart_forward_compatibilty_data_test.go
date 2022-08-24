package commands

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	gatewayMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
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

	integration := setupDeviceDefinitionForSmartCarForwardCompatibility(s.T(), s.pdb, mk, model, year)

	s.mockSmartCarService.EXPECT().GetOrCreateSmartCarIntegration(gomock.Any()).Return(integration.ID, nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncSearchDataCommand{})
	require.NoError(s.T(), err)
	require.NotNilf(s.T(), qryResult, "query result cannot be nil")
	result := qryResult.(SyncSmartCartForwardCompatibilityCommandResult)

	s.NoError(err)
	s.Len(result, 1)
}

func setupDeviceDefinitionForSmartCarForwardCompatibility(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.Integration {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	i := dbtesthelper.SetupCreateSmartCarIntegration(t, pdb)
	dbtesthelper.SetupCreateDeviceIntegration(t, dd, i.ID, pdb)
	return i
}
