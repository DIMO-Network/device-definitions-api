package commands

import (
	"context"
	"os"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elastic"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

	s.queryHandler = NewSyncSearchDataCommandHandler(s.pdb.DBS, s.mockElasticSearch, zerolog.New(os.Stdout))
}

func (s *SyncSearchDataCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *SyncSearchDataCommandHandlerSuite) TestSyncSearchDataCommand_Success() {
	ctx := context.Background()

	model := "Testla"
	mk := "Toyota"
	year := 2020
	metaEngineName := "MetaEngineName"
	engineType := "meta"

	_ = setupDeviceDefinitionForSearchData(s.T(), s.pdb, mk, model, year)

	engineDetail := elastic.EngineDetail{Name: metaEngineName, Type: &engineType}
	getEnginesResp := &elastic.GetEnginesResp{Results: []elastic.EngineDetail{engineDetail}}

	s.mockElasticSearch.EXPECT().GetEngines().Return(getEnginesResp, nil).Times(1)
	s.mockElasticSearch.EXPECT().CreateEngine(gomock.Any(), gomock.Any()).Return(&engineDetail, nil).Times(1)
	s.mockElasticSearch.EXPECT().GetMetaEngineName().Return(metaEngineName).Times(1)
	s.mockElasticSearch.EXPECT().CreateDocumentsBatched(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	s.mockElasticSearch.EXPECT().AddSourceEngineToMetaEngine(gomock.Any(), gomock.Any()).Return(&engineDetail, nil).Times(1)
	s.mockElasticSearch.EXPECT().UpdateSearchSettingsForDeviceDefs(gomock.Any()).Return(nil).Times(2)

	qryResult, err := s.queryHandler.Handle(ctx, &SyncSearchDataCommand{})
	require.NoError(s.T(), err, "handler failed to execute")

	result := qryResult.(SyncSearchDataCommandResult)

	assert.Equal(s.T(), result.Status, true)
}

func setupDeviceDefinitionForSearchData(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
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
