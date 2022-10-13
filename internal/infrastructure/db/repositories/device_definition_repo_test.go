package repositories

import (
	"context"
	_ "embed"

	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

	_ = setupAutoPiIntegration(s.T(), s.pdb)
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
	_ = setupAutoPiIntegration(s.T(), s.pdb)

	dd, err := s.repository.GetOrCreate(ctx, source, mk, model, year)
	s.NoError(err)

	assert.Equal(s.T(), dd.DeviceMakeID, dm.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_Creates_AutoPi_DeviceIntegration() {
	ctx := context.Background()

	model := "Corolla"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	dm := setupDeviceMake(s.T(), s.pdb, mk)
	i := &models.Integration{
		ID:     ksuid.New().String(),
		Type:   models.IntegrationTypeHardware,
		Style:  models.IntegrationStyleAddon,
		Vendor: common.AutoPiVendor,
	}
	s.NoError(i.Insert(ctx, s.pdb.DBS().Writer, boil.Infer()))

	dd, err := s.repository.GetOrCreate(ctx, source, mk, model, year)
	s.NoError(err)
	integration, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(common.AutoPiVendor)).One(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	dis, err := dd.DeviceIntegrations(models.DeviceIntegrationWhere.IntegrationID.EQ(integration.ID)).All(ctx, s.pdb.DBS().Reader)
	s.NoError(err)
	assert.Len(s.T(), dis, 2)

	assert.Equal(s.T(), common.AmericasRegion.String(), dis[0].Region)
	assert.Contains(s.T(), common.EuropeRegion.String(), dis[1].Region)
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

func setupAutoPiIntegration(t *testing.T, pdb db.Store) *models.Integration {
	i := dbtesthelper.SetupCreateAutoPiIntegration(t, pdb)
	return i
}

func Test_slugString(t *testing.T) {

	tests := []struct {
		name string
		term string
		want string
	}{
		{name: "Remove special characters", term: "LÄND ROVER", want: "land-rover"},
		{name: "Remove special characters", term: "Citroën", want: "citroen"},
		{name: "Replace space with dash", term: "Mercedes Benz", want: "mercedes-benz"},
		{name: "All characters lower case", term: "TESLA", want: "tesla"},
		{name: "Replace underscores with a dash", term: "Alfa_Romeo", want: "alfa-romeo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.SlugString(tt.term); got != tt.want {
				t.Errorf("slugString() = %v, want %v", got, tt.want)
			}
		})
	}
}
