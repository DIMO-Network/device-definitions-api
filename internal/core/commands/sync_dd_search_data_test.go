package commands

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

const (
	dbName               = "device_definitions_api"
	migrationsDirRelPath = "../../infrastructure/db/migrations"
)

type SyncSearchDataCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl              *gomock.Controller
	pdb               db.Store
	container         testcontainers.Container
	mockElasticSearch *mocks.MockElasticSearchService
	ctx               context.Context

	queryHandler SyncSearchDataCommandHandler
}

func TestSyncSearchDataCommandHandler(t *testing.T) {
	suite.Run(t, new(SyncSearchDataCommandHandlerSuite))
}

func (s *SyncSearchDataCommandHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockElasticSearch = mocks.NewMockElasticSearchService(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewSyncSearchDataCommandHandler(s.pdb.DBS, s.mockElasticSearch)
}

func (s *SyncSearchDataCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *SyncSearchDataCommandHandlerSuite) TestSyncSearchDataCommand() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020
	metaEngineName := "MetaEngineName"
	engineType := "meta"

	_ = setupDeviceDefinitionForSearchData(s.T(), s.pdb, mk, model, year)

	engineDetail := gateways.EngineDetail{Name: metaEngineName, Type: &engineType}
	getEnginesResp := &gateways.GetEnginesResp{Results: []gateways.EngineDetail{engineDetail}}
	s.mockElasticSearch.EXPECT().GetEngines().Return(getEnginesResp, nil).Times(1)

	s.mockElasticSearch.EXPECT().CreateEngine(gomock.Any(), gomock.Any()).Return(&engineDetail, nil).Times(2)
	s.mockElasticSearch.EXPECT().GetMetaEngineName().Return(metaEngineName).Times(1)
	s.mockElasticSearch.EXPECT().CreateDocumentsBatched(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	s.mockElasticSearch.EXPECT().AddSourceEngineToMetaEngine(gomock.Any(), gomock.Any()).Return(&engineDetail, nil).Times(1)
	s.mockElasticSearch.EXPECT().UpdateSearchSettingsForDeviceDefs(gomock.Any()).Return(nil).Times(2)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncSearchDataCommand{})
	result := qryResult.(SyncSearchDataCommandResult)

	s.NoError(err)
	assert.Equal(s.T(), result.Status, true)
}

func setupDeviceDefinitionForSearchData(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}
