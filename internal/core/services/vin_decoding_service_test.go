package services

import (
	"context"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/testcontainers/testcontainers-go"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type VINDecodingServiceSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                   *gomock.Controller
	pdb                    db.Store
	container              testcontainers.Container
	ctx                    context.Context
	mockDrivlyAPISvc       *mock_gateways.MockDrivlyAPIService
	mockVincarioAPISvc     *mock_gateways.MockVincarioAPIService
	mockAutoIsoAPISvc      *mock_gateways.MockAutoIsoAPIService
	mockDATGroupAPIService *mock_gateways.MockDATGroupAPIService
	mockOnChainSvc         *mock_gateways.MockDeviceDefinitionOnChainService

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
	s.mockAutoIsoAPISvc = mock_gateways.NewMockAutoIsoAPIService(s.ctrl)
	s.mockAutoIsoAPISvc = mock_gateways.NewMockAutoIsoAPIService(s.ctrl)
	s.mockDATGroupAPIService = mock_gateways.NewMockDATGroupAPIService(s.ctrl)
	s.mockOnChainSvc = mock_gateways.NewMockDeviceDefinitionOnChainService(s.ctrl)

	s.vinDecodingService = NewVINDecodingService(s.mockDrivlyAPISvc, s.mockVincarioAPISvc, s.mockAutoIsoAPISvc, dbtesthelper.Logger(),
		s.mockOnChainSvc, s.mockDATGroupAPIService, s.pdb.DBS)
}

func (s *VINDecodingServiceSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *VINDecodingServiceSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func (s *VINDecodingServiceSuite) Test_VINDecodingService_Drivly_Success() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // ford escape 2021
	const makeID = "Ford"
	const country = "US"

	vinInfoResp := &coremodels.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2021",
		Make:                makeID,
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

	result, err := s.vinDecodingService.GetVIN(ctx, vin, dt, coremodels.AllProviders, country)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Source, coremodels.DrivlyProvider)
}

func (s *VINDecodingServiceSuite) Test_VINDecodingService_Tesla() {
	ctx := context.Background()
	const vin = "5YJ3E1EA2PF696023"
	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)
	result, err := s.vinDecodingService.GetVIN(ctx, vin, dt, coremodels.TeslaProvider, "USA")

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Make, "Tesla")
	assert.Equal(s.T(), result.Model, "Model 3")
	assert.Equal(s.T(), string(result.MetaData.JSON), `{"fuel_type":"electric","powertrain_type":"BEV"}`)
	assert.Equal(s.T(), result.Source, coremodels.TeslaProvider)
}

func (s *VINDecodingServiceSuite) Test_VINDecodingService_Vincario_Success() {
	ctx := context.Background()
	const vin = "WAUZZZKM04D018683"
	const makeID = "Test"
	const country = "US"

	vincarioResp := &coremodels.VincarioInfoResponse{
		VIN:                vin,
		ModelYear:          2021,
		Make:               makeID,
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
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(nil, fmt.Errorf("unable to decode"))
	s.mockAutoIsoAPISvc.EXPECT().GetVIN(vin).Times(1).Return(nil, fmt.Errorf("unable to decode"))
	// s.mockDATGroupAPIService.EXPECT().GetVINv2(vin, country).Times(1).Return(nil, fmt.Errorf("unable to decode"))
	// vincario is the last fallback
	s.mockVincarioAPISvc.EXPECT().DecodeVIN(vin).Times(1).Return(vincarioResp, nil)

	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)

	result, err := s.vinDecodingService.GetVIN(ctx, vin, dt, coremodels.AllProviders, country)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Source, coremodels.VincarioProvider)
}

//go:embed autoiso_resp.json
var testAutoIsoJSON []byte

func (s *VINDecodingServiceSuite) Test_VINDecodingService_AutoIso_Success() {
	ctx := context.Background()
	const vin = "WAUZZZKM04D018683"
	const country = "US"

	vinInfoResp := &coremodels.AutoIsoVINResponse{}
	_ = json.Unmarshal(testAutoIsoJSON, vinInfoResp)

	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(nil, fmt.Errorf("unable to decode"))
	s.mockAutoIsoAPISvc.EXPECT().GetVIN(vin).Times(1).Return(vinInfoResp, nil)

	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)

	result, err := s.vinDecodingService.GetVIN(ctx, vin, dt, coremodels.AllProviders, country)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Source, coremodels.AutoIsoProvider)
}

func (s *VINDecodingServiceSuite) Test_VINDecodingService_DD_Default_Success() {
	ctx := context.Background()
	const vin = "0SCZZZ4M0KD018683"
	const country = "US"

	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(s.T(), dm, "Escape", 2020, s.pdb)

	s.mockOnChainSvc.EXPECT().GetDefinitionByID(ctx, dd.ID, gomock.Any()).Times(1).Return(dd, nil, nil)

	result, err := s.vinDecodingService.GetVIN(ctx, vin, dt, coremodels.AllProviders, country)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
}

//go:embed datgroup_resp.xml
var testDATGroupXML []byte

func (s *VINDecodingServiceSuite) Test_VINDecodingService_DATGroup_Success() {
	ctx := context.Background()
	const vin = "ZFADEXTESTSTUB001"
	const country = "TR"

	vinInfoResp := &gateways.DATGroupInfoResponse{
		VIN:               vin,
		MainTypeGroupName: "Test",
		Year:              2023,
	}
	_ = xml.Unmarshal(testDATGroupXML, vinInfoResp)

	s.mockDATGroupAPIService.EXPECT().GetVINv2(vin, country).Times(1).Return(vinInfoResp, nil)

	dt := dbtesthelper.SetupCreateDeviceType(s.T(), s.pdb)

	result, err := s.vinDecodingService.GetVIN(ctx, vin, dt, coremodels.DATGroupProvider, country)

	s.NoError(err)
	assert.Equal(s.T(), result.VIN, vin)
	assert.Equal(s.T(), result.Source, coremodels.DATGroupProvider)
}
