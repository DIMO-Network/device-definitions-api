package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	test_db_helper "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/pkg/dbtest"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/testcontainers/testcontainers-go"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	dbName               = "device_definition_api"
	migrationsDirRelPath = "../../infrastructure/db/migrations"
)

type GetAllIntegrationQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.DbStore
	container testcontainers.Container
	ctx       context.Context

	queryHandler GetAllIntegrationQueryHandler
}

func TestGetAllIntegrationQueryHandler(t *testing.T) {
	suite.Run(t, new(GetAllIntegrationQueryHandlerSuite))
}

func (s *GetAllIntegrationQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = test_db_helper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewGetAllIntegrationQueryHandler(s.pdb.DBS)
}

func (s *GetAllIntegrationQueryHandlerSuite) TearDownTest() {
	test_db_helper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *GetAllIntegrationQueryHandlerSuite) TestGetAllDeviceDefinitionQuery_With_Items() {
	ctx := context.Background()

	vendor := "AutoPI"

	initialData(s.T(), vendor, s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllIntegrationQuery{})
	result := qryResult.([]GetAllIntegrationQueryResult)

	s.NoError(err)
	s.Len(result, 1)
	assert.Equal(s.T(), vendor, result[0].Vendor)
}

func initialData(t *testing.T, vendor string, pdb db.DbStore) models.Integration {
	dm := models.Integration{
		ID:     ksuid.New().String(),
		Type:   "Test",
		Style:  "Test",
		Vendor: vendor,
	}
	err := dm.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "no db error expected")
	return dm
}
