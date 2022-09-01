package repositories

import (
	"context"
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type DeviceIntegrationRepositorySuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	repository DeviceIntegrationRepository
}

func TestDeviceIntegrationRepository(t *testing.T) {
	suite.Run(t, new(DeviceIntegrationRepositorySuite))
}

func (s *DeviceIntegrationRepositorySuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.repository = NewDeviceIntegrationRepository(s.pdb.DBS)
}

func (s *DeviceIntegrationRepositorySuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *DeviceIntegrationRepositorySuite) TestCreateDeviceIntegration__Success() {
	ctx := context.Background()

	deviceDefinitionID := "2D5YSfCcPYW4pTs3NaaqDioUyyl"
	integrationID := "Hummer"
	region := "es-Us"

	di, err := s.repository.Create(ctx, deviceDefinitionID, integrationID, region)

	s.NoError(err)
	assert.Equal(s.T(), di.IntegrationID, integrationID)
}
