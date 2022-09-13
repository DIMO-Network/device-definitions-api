package repositories

import (
	"context"
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type DeviceDefinitionRepositorySuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	repository DeviceDefinitionRepository
}

func TestDeviceDefinitionRepository(t *testing.T) {
	suite.Run(t, new(DeviceDefinitionRepositorySuite))
}

func (s *DeviceDefinitionRepositorySuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.repository = NewDeviceDefinitionRepository(s.pdb.DBS)
}

func (s *DeviceDefinitionRepositorySuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_With_New_Make_Success() {
	ctx := context.Background()

	model := "Murano"
	mk := "Nissan"
	source := "source-01"
	year := 2022

	dd, err := s.repository.GetOrCreate(ctx, source, mk, model, year)

	s.NoError(err)
	assert.Equal(s.T(), dd.Model, model)
	_, err = models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(mk)).One(s.ctx, s.pdb.DBS().Reader)
	assert.NoError(s.T(), err)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_With_Exists_Make_Success() {
	ctx := context.Background()

	model := "Corolla"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	dm := setupDeviceMake(s.T(), s.pdb, mk)

	dd, err := s.repository.GetOrCreate(ctx, source, mk, model, year)

	s.NoError(err)
	assert.Equal(s.T(), dd.DeviceMakeID, dm.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_Existing_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)
	// current logic returns existing DD if duplicate
	dd2, err := s.repository.GetOrCreate(ctx, source, mk, model, year)

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
}

func setupDeviceDefinition(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}

func setupDeviceMake(t *testing.T, pdb db.Store, makeName string) models.DeviceMake {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	return dm
}
