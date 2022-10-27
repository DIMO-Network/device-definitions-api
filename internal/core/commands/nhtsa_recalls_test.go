package commands

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
)

//go:embed test_FLAT_RCL.zip
var TestRecallDataFile []byte

type SyncNHTSARecallsCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                *gomock.Controller
	pdb                 db.Store
	container           testcontainers.Container
	ctx                 context.Context
	recallsRepo         repositories.DeviceNHTSARecallsRepository
	DDRepo              repositories.DeviceDefinitionRepository
	mockRecallsDataFile string
	server              *httptest.Server

	commandHandler SyncNHTSARecallsCommandHandler
}

func TestSyncNHTSARecallsCommandHandler(t *testing.T) {
	suite.Run(t, new(SyncNHTSARecallsCommandHandlerSuite))
}

func (s *SyncNHTSARecallsCommandHandlerSuite) SetupTest() {

	testRemoteFilePath := "/RCL_FLAT.zip"

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != testRemoteFilePath {
			s.T().Errorf("Expected to request '%s', got: %s", testRemoteFilePath, r.URL.Path)
		}
		//if r.Header.Get("Accept") != "application/json" {
		//	s.T().Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		//}
		//w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/x-zip-compressed")
		_, err := w.Write(TestRecallDataFile)
		if err != nil {
			s.T().Errorf("Unable to write server response: %s", err.Error())
		}
	}))

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.recallsRepo = repositories.NewDeviceNHTSARecallsRepository(s.pdb.DBS)
	s.DDRepo = repositories.NewDeviceDefinitionRepository(s.pdb.DBS)
	s.mockRecallsDataFile = s.server.URL + testRemoteFilePath

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	s.commandHandler = NewSyncNHTSARecallsCommandHandler(s.pdb.DBS, &logger, s.recallsRepo, s.DDRepo, &s.mockRecallsDataFile)
}

func (s *SyncNHTSARecallsCommandHandlerSuite) TearDownTest() {
	s.server.Close()
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

type TestMMYSets []struct {
	Make      string
	Model     string
	Year      int
	MakeDef   *models.DeviceMake
	DeviceDef *models.DeviceDefinition
}

func (s *SyncNHTSARecallsCommandHandlerSuite) TestSyncNHTSARecallsCommand_Success() {
	ctx := context.Background()

	testMMY := TestMMYSets{
		{Make: "Ferrari", Model: "F12berlinetta", Year: 2014}, // ALPHANUM CI
		{Make: "Freightliner", Model: "CASCADIA", Year: 2013}, // EXACT
		{Make: "Honda", Model: "Accord Hybrid", Year: 2018},   // EXACT CI
		{Make: "Jaguar", Model: "F-Type", Year: 2021},         // EXACT CI
		{Make: "Mercedes-Benz", Model: "AMG GT", Year: 2019},  // EXACT
		{Make: "Mercedes-Benz", Model: "S-Class", Year: 2005}, // ALPHANUM CI
	}

	dds := setupDeviceDefinitionsForRecallsSync(s.T(), s.pdb, testMMY)
	fmt.Println(dds)

	commandResult, err := s.commandHandler.Handle(ctx, &SyncNHTSARecallsCommand{})
	s.NoError(err)
	s.Assert().Equal(SyncNHTSARecallsCommandResult{
		InsertedCount: 8,
		MatchCount:    8,
	}, commandResult)

	// Expecting 8 rows, 2 of each matchType
	recalls, err := models.DeviceNhtsaRecalls().Count(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	s.Assert().Equal(int64(8), recalls)
	recalls, err = models.DeviceNhtsaRecalls(
		qm.Where("(metadata ->> 'matchType') = 'EXACT'"),
	).Count(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	s.Assert().Equal(int64(2), recalls)
	recalls, err = models.DeviceNhtsaRecalls(
		qm.Where("(metadata ->> 'matchType') = 'EXACT CI'"),
	).Count(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	s.Assert().Equal(int64(2), recalls)
	recalls, err = models.DeviceNhtsaRecalls(
		qm.Where("(metadata ->> 'matchType') = 'ALPHANUM CI'"),
	).Count(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	s.Assert().Equal(int64(2), recalls)
	recalls, err = models.DeviceNhtsaRecalls(
		qm.Where("(metadata ->> 'matchType') = 'NONE'"),
	).Count(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	s.Assert().Equal(int64(2), recalls)

}

func setupDeviceDefinitionsForRecallsSync(t *testing.T, pdb db.Store, mmySet TestMMYSets) []*models.DeviceDefinition {
	dds := make([]*models.DeviceDefinition, len(mmySet))
	for i, mmy := range mmySet {
		for j := 0; j < i; j++ {
			if mmySet[j].MakeDef != nil && mmySet[j].MakeDef.Name == mmy.Make {
				mmySet[i].MakeDef = mmySet[j].MakeDef
				break
			}
		}
		if mmySet[i].MakeDef == nil {
			dm := dbtesthelper.SetupCreateMake(t, mmy.Make, pdb)
			mmySet[i].MakeDef = &dm
		}
		mmySet[i].DeviceDef = dbtesthelper.SetupCreateDeviceDefinition(t, *mmySet[i].MakeDef, mmy.Model, mmy.Year, pdb)
		dds[i] = mmySet[i].DeviceDef
	}
	return dds
}
