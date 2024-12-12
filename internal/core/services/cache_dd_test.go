package services

import (
	"context"
	_ "embed"
	"fmt"
	mock_repository "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type CacheDeviceDefinitionSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                               *gomock.Controller
	pdb                                db.Store
	container                          testcontainers.Container
	repository                         repositories.DeviceDefinitionRepository
	mockRedis                          *mocks.MockRedisCacheService
	ctx                                context.Context
	mockDeviceDefinitionOnChainService *mock_gateways.MockDeviceDefinitionOnChainService

	cache               DeviceDefinitionCacheService
	mockDeviceMakesRepo *mock_repository.MockDeviceMakeRepository
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
	s.mockDeviceDefinitionOnChainService = mock_gateways.NewMockDeviceDefinitionOnChainService(s.ctrl)
	s.mockDeviceMakesRepo = mock_repository.NewMockDeviceMakeRepository(s.ctrl)

	s.repository = repositories.NewDeviceDefinitionRepository(s.pdb.DBS, s.mockDeviceDefinitionOnChainService)
	s.cache = NewDeviceDefinitionCacheService(s.mockRedis, s.repository, s.mockDeviceMakesRepo, s.mockDeviceDefinitionOnChainService, nil)
}

func (s *CacheDeviceDefinitionSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *CacheDeviceDefinitionSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func (s *CacheDeviceDefinitionSuite) TestCacheDeviceDefinitionByID_Success() {
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

func (s *CacheDeviceDefinitionSuite) TestCacheDeviceDefinitionByMMY_Success() {
	ctx := context.Background()

	model := "Hummer"
	mk := "Toyota"
	year := 2022

	dd := setupDeviceDefinition(s.T(), s.pdb, mk, model, year)

	s.mockRedis.EXPECT().Get(ctx, gomock.Any()).Times(1)
	s.mockRedis.EXPECT().Set(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	result, err := s.cache.GetDeviceDefinitionByMakeModelAndYears(ctx, mk, model, year)

	s.NoError(err)
	assert.Equal(s.T(), result.DeviceDefinitionID, dd.ID)
}

func setupDeviceDefinition(t *testing.T, pdb db.Store, makeName string, modelName string, year int) *models.DeviceDefinition {
	dm := dbtesthelper.SetupCreateMake(t, makeName, pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, modelName, year, pdb)
	return dd
}
