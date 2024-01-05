package queries

import (
	"context"
	"crypto/sha1"
	"fmt"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type GetRecallsByMakeQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	queryHandler GetRecallsByMakeQueryHandler
}

func TestGetRecallsByMakeQueryHandler(t *testing.T) {
	suite.Run(t, new(GetRecallsByMakeQueryHandlerSuite))
}

func (s *GetRecallsByMakeQueryHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewGetRecallsByMakeQueryHandler(s.pdb.DBS)
}

func (s *GetRecallsByMakeQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *GetRecallsByMakeQueryHandlerSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func (s *GetRecallsByMakeQueryHandlerSuite) TestGetRecallsByMakeQuery_Success() {
	ctx := context.Background()

	mk := "Toyota"
	model1 := "Hilux"
	model2 := "Prado"

	dm := setupDeviceMake(s.T(), mk, s.pdb)
	_ = setupDeviceDefinitionWithNhtsa(s.T(), dm, model1, cutoffYear, s.pdb)
	_ = setupDeviceDefinitionWithNhtsa(s.T(), dm, model2, cutoffYear, s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetRecallsByMakeQuery{
		MakeID: dm.ID,
	})
	result := qryResult.(*p_grpc.GetRecallsResponse)

	s.NoError(err)
	s.Len(result.Recalls, 2)
}

func (s *GetRecallsByMakeQueryHandlerSuite) TestGetRecallsByMakeQuery_With_Empty_Success() {
	ctx := context.Background()

	mk := "Toyota"

	dm := setupDeviceMake(s.T(), mk, s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetRecallsByMakeQuery{
		MakeID: dm.ID,
	})
	result := qryResult.(*p_grpc.GetRecallsResponse)

	s.NoError(err)
	s.Len(result.Recalls, 0)
}

func setupDeviceMake(t *testing.T, makeName string, pdb db.Store) models.DeviceMake {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)

	return dm
}

func setupDeviceDefinitionWithNhtsa(t *testing.T, dm models.DeviceMake, modelName string, year int, pdb db.Store) *models.DeviceDefinition {
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)

	recall := &models.DeviceNhtsaRecall{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: null.StringFrom(dd.ID),
		DataYeartxt:        year,
		DataDescDefect:     fmt.Sprintf("description %s %d", modelName, year),
	}
	hasher := sha1.New()
	hasher.Write([]byte(recall.ID + recall.DataDescDefect))
	recall.Hash = hasher.Sum(nil)

	err := recall.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm

	return dd
}
