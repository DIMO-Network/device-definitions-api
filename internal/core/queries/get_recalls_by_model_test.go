package queries

import (
	"context"
	"testing"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type GetRecallsByModelQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	queryHandler GetRecallsByModelQueryHandler
}

func TestGetRecallsByModelQueryHandler(t *testing.T) {
	suite.Run(t, new(GetRecallsByModelQueryHandlerSuite))
}

func (s *GetRecallsByModelQueryHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewGetRecallsByModelQueryHandler(s.pdb.DBS)
}

func (s *GetRecallsByModelQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *GetRecallsByModelQueryHandlerSuite) TestGetRecallsByModelQuery_Success() {
	ctx := context.Background()

	mk := "Toyota"
	model1 := "Hilux"

	dm := setupDeviceMake(s.T(), mk, s.pdb)
	dd := setupDeviceDefinitionWithNhtsa(s.T(), dm, model1, int(cutoffYear), s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetRecallsByModelQuery{
		DeviceDefinitionID: dd.ID,
	})
	result := qryResult.(*p_grpc.GetRecallsResponse)

	s.NoError(err)
	s.Len(result.Recalls, 1)
}

func (s *GetRecallsByModelQueryHandlerSuite) TestGetRecallsByModelQuery_With_Empty_Success() {
	ctx := context.Background()

	qryResult, err := s.queryHandler.Handle(ctx, &GetRecallsByModelQuery{
		DeviceDefinitionID: "dm.ID",
	})
	result := qryResult.(*p_grpc.GetRecallsResponse)

	s.NoError(err)
	s.Len(result.Recalls, 0)
}
