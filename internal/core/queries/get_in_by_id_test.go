package queries

import (
	"context"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type GetIntegrationByIDQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	queryHandler GetIntegrationByIDQueryHandler
}

func TestGetIntegrationByIDQueryHandler(t *testing.T) {
	suite.Run(t, new(GetIntegrationByIDQueryHandlerSuite))
}

func (s *GetIntegrationByIDQueryHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewGetIntegrationByIDQueryHandler(s.pdb.DBS)
}

func (s *GetIntegrationByIDQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *GetIntegrationByIDQueryHandlerSuite) TestGetIntegrationByIDQuery_Success() {
	ctx := context.Background()

	integration := setupCreateSmartCarIntegration(s.T(), s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetIntegrationByIDQuery{
		IntegrationID: integration.ID,
	})
	result := qryResult.(coremodels.GetIntegrationQueryResult)

	s.NoError(err)
	assert.Equal(s.T(), integration.ID, result.ID)
	assert.Equal(s.T(), integration.Type, result.Type)
	assert.Equal(s.T(), integration.Vendor, result.Vendor)
	assert.Equal(s.T(), integration.TokenID.Int, result.TokenID)
}

func (s *GetIntegrationByIDQueryHandlerSuite) TestGetIntegrationByIDQuery_Exception() {
	ctx := context.Background()

	integrationID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"

	qryResult, err := s.queryHandler.Handle(ctx, &GetIntegrationByIDQuery{
		IntegrationID: integrationID,
	})

	s.Nil(qryResult)
	s.Error(err)
}
