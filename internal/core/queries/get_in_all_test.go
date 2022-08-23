package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const (
	dbName               = "device_definitions_api"
	migrationsDirRelPath = "../../infrastructure/db/migrations"
)

type GetAllIntegrationQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	queryHandler GetAllIntegrationQueryHandler
}

func TestGetAllIntegrationQueryHandler(t *testing.T) {
	suite.Run(t, new(GetAllIntegrationQueryHandlerSuite))
}

func (s *GetAllIntegrationQueryHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewGetAllIntegrationQueryHandler(s.pdb.DBS)
}

func (s *GetAllIntegrationQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *GetAllIntegrationQueryHandlerSuite) TestGetAllDeviceDefinitionQuery_With_Items() {
	ctx := context.Background()

	integration := setupCreateSmartCarIntegration(s.T(), s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllIntegrationQuery{})
	result := qryResult.([]GetAllIntegrationQueryResult)

	s.NoError(err)
	s.Len(result, 1)
	assert.Equal(s.T(), integration.Vendor, result[0].Vendor)
}

func setupCreateSmartCarIntegration(t *testing.T, pdb db.Store) models.Integration {
	integration := models.Integration{
		ID:               ksuid.New().String(),
		Type:             models.IntegrationTypeAPI,
		Style:            models.IntegrationStyleWebhook,
		Vendor:           "SmartCar",
		RefreshLimitSecs: 1800,
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return integration
}
