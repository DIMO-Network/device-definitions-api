package repositories

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/shared"

	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

type DeviceDefinitionRepositorySuite struct {
	suite.Suite
	*require.Assertions

	ctrl                               *gomock.Controller
	pdb                                db.Store
	container                          testcontainers.Container
	ctx                                context.Context
	mockDeviceDefinitionOnChainService *mock_gateways.MockDeviceDefinitionOnChainService

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

	s.mockDeviceDefinitionOnChainService = mock_gateways.NewMockDeviceDefinitionOnChainService(s.ctrl)

	s.repository = NewDeviceDefinitionRepository(s.pdb.DBS, s.mockDeviceDefinitionOnChainService)
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
	hardwareTemplateID := ksuid.New().String()

	_ = setupAutoPiIntegration(s.T(), s.pdb)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	dd, err := s.repository.GetOrCreate(ctx, nil, source, "", mk, model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)

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
	hardwareTemplateID := ksuid.New().String()

	dm := setupDeviceMake(s.T(), s.pdb, mk)
	_ = setupAutoPiIntegration(s.T(), s.pdb)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	dd, err := s.repository.GetOrCreate(ctx, nil, source, "", mk, model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)
	s.NoError(err)

	assert.Equal(s.T(), dd.DeviceMakeID, dm.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_With_Exists_Make_SuccessById() {
	ctx := context.Background()

	model := "Corolla"
	source := "source-01"
	year := 2022
	hardwareTemplateID := ksuid.New().String()

	dm := setupDeviceMake(s.T(), s.pdb, "Toyota")
	_ = setupAutoPiIntegration(s.T(), s.pdb)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	dd, err := s.repository.GetOrCreate(ctx, nil, source, "", dm.ID, model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)
	s.NoError(err)

	assert.Equal(s.T(), dd.DeviceMakeID, dm.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_With_InvalidMakeId_err() {
	ctx := context.Background()

	model := "Corolla"
	source := "source-01"
	year := 2022
	hardwareTemplateID := ksuid.New().String()

	_ = setupDeviceMake(s.T(), s.pdb, "Toyota")
	_ = setupAutoPiIntegration(s.T(), s.pdb)

	dd, err := s.repository.GetOrCreate(ctx, nil, source, "", ksuid.New().String(), model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)
	s.Error(err, "expected an error")
	s.Assert().Nil(dd)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_Creates_AutoPi_DeviceIntegration() {
	ctx := context.Background()

	model := "Corolla"
	mk := "Toyota"
	source := "source-01"
	year := 2022
	hardwareTemplateID := ksuid.New().String()

	dm := setupDeviceMake(s.T(), s.pdb, mk)
	i := &models.Integration{
		ID:     ksuid.New().String(),
		Type:   models.IntegrationTypeHardware,
		Style:  models.IntegrationStyleAddon,
		Vendor: common.AutoPiVendor,
	}
	s.NoError(i.Insert(ctx, s.pdb.DBS().Writer, boil.Infer()))

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	dd, err := s.repository.GetOrCreate(ctx, nil, source, "", mk, model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)
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
	hardwareTemplateID := ksuid.New().String()

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)
	// current logic returns existing DD if duplicate
	dd2, err := s.repository.GetOrCreate(ctx, nil, source, "", mk, model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateOrUpdateDeviceDefinition_New_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	source := "source-01"
	year := 2022

	dm := setupDeviceMake(s.T(), s.pdb, mk)
	dd := &models.DeviceDefinition{
		ID:           ksuid.New().String(),
		DeviceMakeID: dm.ID,
		Model:        model,
		Source:       null.StringFrom(source),
		Year:         int16(year),
		Verified:     false,
		ModelSlug:    shared.SlugString(model),
	}

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)
	dd2, err := s.repository.CreateOrUpdate(ctx, dd, []*models.DeviceStyle{}, []*models.DeviceIntegration{})

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateOrUpdateDeviceDefinition_Existing_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)

	newModel := "Hulix Pro"
	newYear := 2023
	newSource := "source-02"

	dd.Model = newModel
	dd.Year = int16(newYear)
	dd.Source = null.StringFrom(newSource)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)
	dd2, err := s.repository.CreateOrUpdate(ctx, dd, []*models.DeviceStyle{}, []*models.DeviceIntegration{})

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
	assert.Equal(s.T(), dd.Model, dd2.Model)
	assert.Equal(s.T(), dd.Year, dd2.Year)
	assert.Equal(s.T(), dd.Source, dd2.Source)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateOrUpdateDeviceDefinition_With_NewStyles_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinitionWithStyles(s.T(), s.pdb, mk, model, year)

	newStyles := []*models.DeviceStyle{}

	for _, style := range dd.R.DeviceStyles {
		newStyles = append(newStyles, style)
	}

	// add new style
	newStyles = append(newStyles, &models.DeviceStyle{
		ID:                 ksuid.New().String(),
		Name:               "New Style",
		DeviceDefinitionID: dd.ID,
		Source:             "New Source",
		SubModel:           "New SubModel",
		ExternalStyleID:    ksuid.New().String(),
	})

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	dd2, err := s.repository.CreateOrUpdate(ctx, dd, newStyles, []*models.DeviceIntegration{})

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
	assert.Equal(s.T(), dd.Model, dd2.Model)
	assert.Equal(s.T(), dd.Year, dd2.Year)
	assert.Equal(s.T(), dd.Source, dd2.Source)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateOrUpdateDeviceDefinition_With_NewIntegration_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	i := setupIntegrationForDeviceIntegration(s.T(), s.pdb)
	dd := setupDeviceDefinitionWithIntegrations(s.T(), s.pdb, mk, model, year)

	newDeviceIntegrations := []*models.DeviceIntegration{}

	for _, integration := range dd.R.DeviceIntegrations {
		newDeviceIntegrations = append(newDeviceIntegrations, integration)
	}

	// add new integrations
	newDeviceIntegrations = append(newDeviceIntegrations, &models.DeviceIntegration{
		IntegrationID:      i.ID,
		DeviceDefinitionID: dd.ID,
		Region:             "east-us",
	})

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	dd2, err := s.repository.CreateOrUpdate(ctx, dd, []*models.DeviceStyle{}, newDeviceIntegrations)

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestCreateOrUpdateDeviceDefinition_With_Vehicle_DeviceTypes_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)
	dt, _ := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(dd.DeviceTypeID.String)).One(ctx, s.pdb.DBS().Reader)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	deviceTypeInfo := make(map[string]interface{})
	metaData := make(map[string]interface{})
	var ai map[string][]interface{}
	defaultValue := "defaultValue"
	if err := dt.Properties.Unmarshal(&ai); err == nil {
		metaData["mpg"] = defaultValue
	}
	deviceTypeInfo[dt.Metadatakey] = metaData
	j, _ := json.Marshal(deviceTypeInfo)
	dd.Metadata = null.JSONFrom(j)
	// current logic returns existing DD if duplicate
	dd2, err := s.repository.CreateOrUpdate(ctx, dd, []*models.DeviceStyle{}, []*models.DeviceIntegration{})

	s.NoError(err)
	assert.Equal(s.T(), dd.ID, dd2.ID)
}

func (s *DeviceDefinitionRepositorySuite) TestGetDeviceDefinition_By_Slug_Success() {
	ctx := context.Background()

	model := "Hilux"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)
	// current logic returns existing DD if duplicate
	dd2, err := s.repository.GetBySlugAndYear(ctx, dd.ModelSlug, year, true)

	s.NoError(err)
	assert.Equal(s.T(), dd2.ID, dd.ID)
	assert.Equal(s.T(), dd2.R.DeviceMake.Name, mk)
	assert.Equal(s.T(), dd2.Year, int16(year))
}

func (s *DeviceDefinitionRepositorySuite) TestGetDeviceDefinition_Nil_By_Slug_Success() {
	ctx := context.Background()

	mk := "Toyota"
	year := 2022

	dd, _ := s.repository.GetBySlugAndYear(ctx, mk, year, true)

	s.Nil(dd)
}

func setupDeviceDefinition(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)

	return dd
}

func setupDeviceDefinitionWithStyles(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)

	ds1 := dbtesthelper.SetupCreateStyle(t, dd.ID, "4dr SUV 4WD", "edmunds", "Wagon", pdb)
	ds2 := dbtesthelper.SetupCreateStyle(t, dd.ID, "Hard Top 2dr SUV AWD", "edmunds", "Open Top", pdb)

	dd.R = dd.R.NewStruct()
	dd.R.DeviceStyles = append(dd.R.DeviceStyles, &ds1)
	dd.R.DeviceStyles = append(dd.R.DeviceStyles, &ds2)
	dd.R.DeviceMake = &dm

	return dd
}

func setupDeviceDefinitionWithIntegrations(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)

	i := dbtesthelper.SetupCreateHardwareIntegration(t, pdb)
	di := dbtesthelper.SetupCreateDeviceIntegration(t, dd, i.ID, "Americas", pdb)

	dd.R = dd.R.NewStruct()
	dd.R.DeviceIntegrations = append(dd.R.DeviceIntegrations, di)
	dd.R.DeviceMake = &dm

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
			if got := shared.SlugString(tt.term); got != tt.want {
				t.Errorf("slugString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func (s *DeviceDefinitionRepositorySuite) TestCreateDeviceDefinition_With_Make_Empty_Error() {
	ctx := context.Background()

	model := "Corolla"
	mk := ""
	source := "source-01"
	year := 2022
	hardwareTemplateID := ksuid.New().String()

	_, err := s.repository.GetOrCreate(ctx, nil, source, "", mk, model, year, "vehicle", null.JSON{}, false, &hardwareTemplateID)
	s.Error(err)
}
