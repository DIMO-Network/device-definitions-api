package commands

import (
	"context"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"

	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type SyncVinNumbersSuite struct {
	suite.Suite
	*require.Assertions

	ctrl               *gomock.Controller
	pdb                db.Store
	container          testcontainers.Container
	ctx                context.Context
	mockDrivlyAPISvc   *mock_gateways.MockDrivlyAPIService
	mockVincarioAPISvc *mock_gateways.MockVincarioAPIService
	mockVINService     *mock_services.MockVINDecodingService

	commandHandler SyncVinNumbersCommandHandler
}

func TestSyncVinNumbersCommandHandler(t *testing.T) {
	suite.Run(t, new(SyncVinNumbersSuite))
}

func (s *SyncVinNumbersSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.ctx = context.Background()

	s.mockDrivlyAPISvc = mock_gateways.NewMockDrivlyAPIService(s.ctrl)
	s.mockVincarioAPISvc = mock_gateways.NewMockVincarioAPIService(s.ctrl)
	s.mockVINService = mock_services.NewMockVINDecodingService(s.ctrl)

	repo := repositories.NewDeviceDefinitionRepository(s.pdb.DBS)
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)
	s.commandHandler = NewSyncVinNumbersCommand(s.pdb.DBS, s.mockVINService, repo, dbtesthelper.Logger())
}

func (s *SyncVinNumbersSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *SyncVinNumbersSuite) TestHandle_Success_Sync_Drivly_WithExisting() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	_ = dbtesthelper.SetupCreateDeviceDefinition(s.T(), dm, "Escape", 2021, s.pdb)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: "XLE Hybrid",
	}

	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.commandHandler.Handle(s.ctx, &SyncVinNumbersCommand{VINNumbers: []string{vin}})
	s.NoError(err)
	result := qryResult.(*SyncVinNumbersCommandResult)
	s.Assert().Equal(true, result.Status)
}

func (s *SyncVinNumbersSuite) TestHandle_Success_Sync_Drivly_CreatesDD() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021
	const wmi = "1FM"

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateWMI(s.T(), wmi, dm.ID, s.pdb)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: "XLE Hybrid",
	}

	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.commandHandler.Handle(s.ctx, &SyncVinNumbersCommand{VINNumbers: []string{vin}})
	s.NoError(err)
	result := qryResult.(*SyncVinNumbersCommandResult)

	s.Assert().Equal(true, result.Status)
}

func (s *SyncVinNumbersSuite) TestHandle_Success_Sync_Drivly_WithExistingWMI() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	_ = dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)
	wmi := models.Wmi{
		Wmi:          "1FM",
		DeviceMakeID: dm.ID,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: "XLE Hybrid",
	}

	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.commandHandler.Handle(s.ctx, &SyncVinNumbersCommand{VINNumbers: []string{vin}})
	s.NoError(err)
	result := qryResult.(*SyncVinNumbersCommandResult)
	s.Assert().Equal(true, result.Status)
}
