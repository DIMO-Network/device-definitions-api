package commands

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	repositoryMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	gatewayMock "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

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

	_ = setupDeviceDefinitionForSearchData(s.T(), s.pdb, mk, model, year)

	smartCarCompatibilityData := &gateways.SmartCarCompatibilityData{}

	s.mockSmartCarService.EXPECT().GetSmartCarVehicleData().Return(smartCarCompatibilityData, nil).Times(1)
	s.mockSmartCarService.EXPECT().GetOrCreateSmartCarIntegration(gomock.Any()).Return(gomock.Any(), nil).Times(1)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncSearchDataCommand{})
	result := qryResult.(SyncSearchDataCommandResult)

	s.NoError(err)
	s.NotNil(result)
}

func setupDeviceDefinitionForSmartCarCompatibility(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {

	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}
