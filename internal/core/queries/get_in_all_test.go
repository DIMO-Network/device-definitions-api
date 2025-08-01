package queries

import (
	"context"
	"fmt"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
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

func (s *GetAllIntegrationQueryHandlerSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func (s *GetAllIntegrationQueryHandlerSuite) TestGetAllDeviceDefinitionQuery_With_Items_Success() {
	ctx := context.Background()

	integration := setupCreateSmartCarIntegration(s.T(), s.pdb)

	qryResult, err := s.queryHandler.Handle(ctx, &GetAllIntegrationQuery{})
	result := qryResult.([]coremodels.GetIntegrationQueryResult)

	s.NoError(err)
	s.Len(result, 1)
	assert.Equal(s.T(), integration.Vendor, result[0].Vendor)
	assert.Equal(s.T(), integration.TokenID.Int, result[0].TokenID)
}

func setupCreateSmartCarIntegration(t *testing.T, pdb db.Store) models.Integration {
	integration := models.Integration{
		ID:               ksuid.New().String(),
		Type:             models.IntegrationTypeAPI,
		Style:            models.IntegrationStyleWebhook,
		Vendor:           "SmartCar",
		RefreshLimitSecs: 1800,
		TokenID:          null.IntFrom(1),
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	assert.NoError(t, err, "database error")
	return integration
}
