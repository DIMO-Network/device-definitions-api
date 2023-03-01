package services

import (
	"context"
	_ "embed"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/testcontainers/testcontainers-go"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VINDecodingServiceSuite struct {
	suite.Suite
	*require.Assertions

	ctrl               *gomock.Controller
	pdb                db.Store
	container          testcontainers.Container
	ctx                context.Context
	mockDrivlyAPISvc   *mock_gateways.MockDrivlyAPIService
	mockVincarioAPISvc *mock_gateways.MockVincarioAPIService

	vinDecodingService VINDecodingService
}

func TestVINDecodingService(t *testing.T) {
	suite.Run(t, new(VINDecodingServiceSuite))
}

func (s *VINDecodingServiceSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.mockDrivlyAPISvc = mock_gateways.NewMockDrivlyAPIService(s.ctrl)
	s.mockVincarioAPISvc = mock_gateways.NewMockVincarioAPIService(s.ctrl)

	s.vinDecodingService = NewVINDecodingService(s.mockDrivlyAPISvc, s.mockVincarioAPISvc, dbtesthelper.Logger())
}

func (s *VINDecodingServiceSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *VINDecodingServiceSuite) Test_VINDecodingService_Drivly_Success() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021
	const make = "Ford"

	vinInfoResp := &gateways.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2021",
		Make:                make,
		Model:               "Escape",
		SubModel:            "Hybrid",
		Trim:                "XLE",
		Generation:          4,
		ManufacturerCode:    "1234",
		Drive:               "AWD",
		Engine:              "4 Cyl",
		EngineDetails:       "16-Valve, Inline-4, GDI, Hybrid, DOHC, Atkinson Cycle 2.5 L",
		Doors:               4,
		MsrpBase:            23000,
		Fuel:                "Hybrid",
		FuelTankCapacityGal: 15.5,
		Mpg:                 25,
		MpgCity:             21,
		MpgHighway:          31,
		Wheelbase:           "106 WB",
	}
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(vinInfoResp, nil)

	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)

	result, err := s.vinDecodingService.GetVIN(vin, dt, coremodels.AllProviders)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Source, coremodels.DrivlyProvider)
}

func (s *VINDecodingServiceSuite) Test_VINDecodingService_Vincario_Success() {
	const vin = "WAUZZZ4M0KD018683"
	const make = "Test"

	vinInfoResp := &gateways.VincarioInfoResponse{
		VIN:                vin,
		ModelYear:          2021,
		Make:               make,
		Model:              "Escape",
		EngineManufacturer: "1234",
		Drive:              "AWD",
		EngineCode:         "4 Cyl",
		EngineType:         "16-Valve, Inline-4, GDI, Hybrid, DOHC, Atkinson Cycle 2.5 L",
		NumberOfDoors:      4,
		FuelType:           "Hybrid",
		Height:             31,
		Wheelbase:          1,
	}
	s.mockVincarioAPISvc.EXPECT().DecodeVIN(vin).Times(1).Return(vinInfoResp, nil)
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(nil, nil)

	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)

	result, err := s.vinDecodingService.GetVIN(vin, dt, coremodels.AllProviders)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Source, coremodels.VincarioProvider)
}
