package services

import (
	"context"
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/pkg/dbtest"
	"github.com/testcontainers/testcontainers-go"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CacheDeviceDefinitionSuite struct {
	suite.Suite
	*require.Assertions

	ctrl       *gomock.Controller
	pdb        db.Store
	container  testcontainers.Container
	repository repositories.DeviceDefinitionRepository
	mockRedis  *mocks.MockRedisCacheService
	ctx        context.Context

	cache DeviceDefinitionCacheService
}

func TestCacheDeviceDefinition(t *testing.T) {
	suite.Run(t, new(CacheDeviceDefinitionSuite))
}

func (s *CacheDeviceDefinitionSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.mockRedis = mocks.NewMockRedisCacheService(s.ctrl)

	s.repository = repositories.NewDeviceDefinitionRepository(s.pdb.DBS)
	s.cache = NewDeviceDefinitionCacheService(s.mockRedis, s.repository)
}

func (s *CacheDeviceDefinitionSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CacheDeviceDefinitionSuite) TestCacheDeviceDefinition_Success() {
	ctx := context.Background()

	model := "Hummer"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)

	s.mockRedis.EXPECT().Get(ctx, gomock.Any()).Times(1)
	s.mockRedis.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	result, err := s.cache.GetDeviceDefinitionByID(ctx, dd.ID)

	s.NoError(err)
	assert.Equal(s.T(), result.DeviceDefinitionID, dd.ID)
}

func setupDeviceDefinition(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}
