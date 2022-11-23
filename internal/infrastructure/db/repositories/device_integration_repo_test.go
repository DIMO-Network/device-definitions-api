package repositories

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
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

func (s *DeviceIntegrationRepositorySuite) TestCreateDeviceIntegration_Success() {
	ctx := context.Background()

	region := "es-Us"

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinitionForDeviceIntegration(s.T(), s.pdb, mk, model, year)
	i := setupIntegrationForDeviceIntegration(s.T(), s.pdb)

	featureArray := `[
	  {
		"featureKey": "fuel_type",
		"supportLevel": 0
	  }
	]`

	var metaData []map[string]interface{}
	json.Unmarshal([]byte(featureArray), &metaData)

	di, err := s.repository.Create(ctx, dd.ID, i.ID, region, metaData)

	s.NoError(err)
	assert.Equal(s.T(), di.IntegrationID, i.ID)
}

func (s *DeviceIntegrationRepositorySuite) TestCreateDeviceIntegration_Exception() {
	ctx := context.Background()

	region := "es-Us"

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinitionForDeviceIntegration(s.T(), s.pdb, mk, model, year)

	di, err := s.repository.Create(ctx, dd.ID, "integration-ID", region, nil)

	s.Nil(di)
	s.Error(err)
}

func setupDeviceDefinitionForDeviceIntegration(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}

func setupIntegrationForDeviceIntegration(t *testing.T, pdb db.Store) *models.Integration {
	i := dbtesthelper.SetupCreateSmartCarIntegration(t, pdb)
	return i
}
