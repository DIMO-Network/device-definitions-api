package queries

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared"
	"github.com/volatiletech/null/v8"

	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DecodeVINQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                      *gomock.Controller
	pdb                       db.Store
	container                 testcontainers.Container
	ctx                       context.Context
	mockDrivlyAPISvc          *mock_gateways.MockDrivlyAPIService
	mockVincarioAPISvc        *mock_gateways.MockVincarioAPIService
	mockVINService            *mock_services.MockVINDecodingService
	mockFuelAPIService        *mock_gateways.MockFuelAPIService
	mockPowerTrainTypeService *mock_services.MockPowerTrainTypeService

	queryHandler DecodeVINQueryHandler
}

func TestDecodeVINQueryHandler(t *testing.T) {
	suite.Run(t, new(DecodeVINQueryHandlerSuite))
}

func (s *DecodeVINQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.ctx = context.Background()

	s.mockDrivlyAPISvc = mock_gateways.NewMockDrivlyAPIService(s.ctrl)
	s.mockVincarioAPISvc = mock_gateways.NewMockVincarioAPIService(s.ctrl)
	s.mockVINService = mock_services.NewMockVINDecodingService(s.ctrl)
	s.mockPowerTrainTypeService = mock_services.NewMockPowerTrainTypeService(s.ctrl)
	repo := repositories.NewDeviceDefinitionRepository(s.pdb.DBS)
	vinRepository := repositories.NewVINRepository(s.pdb.DBS)
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)
	s.queryHandler = NewDecodeVINQueryHandler(s.pdb.DBS, s.mockVINService, vinRepository, repo, dbtesthelper.Logger(), s.mockFuelAPIService, s.mockPowerTrainTypeService)
}

func (s *DecodeVINQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingDD_UpdatesAttributes_CreatesStyle() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(s.T(), dm, "Escape", 2021, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &gateways.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2021",
		Make:                dm.Name,
		Model:               dd.Model,
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

	deviceTypeInfo := make(map[string]interface{})
	deviceTypeInfo["mpg_city"] = vinInfoResp.MpgCity
	deviceTypeInfo["mpg_highway"] = vinInfoResp.MpgHighway
	deviceTypeInfo["mpg"] = vinInfoResp.Mpg
	deviceTypeInfo["base_msrp"] = vinInfoResp.MsrpBase
	deviceTypeInfo["fuel_tank_capacity_gal"] = vinInfoResp.FuelTankCapacityGal
	deviceTypeInfo["fuel_type"] = vinInfoResp.Fuel
	deviceTypeInfo["wheelbase"] = vinInfoResp.Wheelbase
	deviceTypeInfo["generation"] = vinInfoResp.Generation
	deviceTypeInfo["number_of_doors"] = vinInfoResp.Doors
	deviceTypeInfo["manufacturer_code"] = vinInfoResp.ManufacturerCode
	deviceTypeInfo["driven_wheels"] = vinInfoResp.Drive

	yr, _ := strconv.Atoi(vinInfoResp.Year)
	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: buildStyleName(vinInfoResp),
		SubModel:  vinInfoResp.SubModel,
		Source:    "drivly",
		Year:      int32(yr),
		Make:      vinInfoResp.Make,
		Model:     vinInfoResp.Model,
	}

	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)

	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)
	// db setup

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DeviceDefinitionId)
	s.Assert().Equal(dm.ID, result.DeviceMakeId)
	// validate WMI was inserted
	wmi, err := models.Wmis().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal("1FM", wmi.Wmi)
	s.Assert().Equal(dm.ID, wmi.DeviceMakeID)
	// validate style was created
	ds, err := models.DeviceStyles().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)
	s.Assert().Equal(vinInfoResp.Trim+" "+vinInfoResp.SubModel, ds.Name)
	s.Assert().Equal(vinInfoResp.SubModel, ds.SubModel)
	s.Assert().Equal("drivly", ds.Source)
	s.Assert().Equal(ds.ExternalStyleID, common.SlugString(vinInfoResp.Trim+" "+vinInfoResp.SubModel))

	// validate metadata was updated on DD
	ddUpdated, err := models.DeviceDefinitions().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)

	assert.Equal(s.T(), vinInfoResp.Wheelbase, gjson.GetBytes(ddUpdated.Metadata.JSON, "vehicle_info.wheelbase").String())
	assert.Equal(s.T(), int64(vinInfoResp.Doors), gjson.GetBytes(ddUpdated.Metadata.JSON, "vehicle_info.number_of_doors").Int())
	assert.Equal(s.T(), int64(vinInfoResp.MsrpBase), gjson.GetBytes(ddUpdated.Metadata.JSON, "vehicle_info.base_msrp").Int())
	assert.Equal(s.T(), int64(vinInfoResp.Mpg), gjson.GetBytes(ddUpdated.Metadata.JSON, "vehicle_info.mpg").Int())
	assert.Equal(s.T(), vinInfoResp.FuelTankCapacityGal, gjson.GetBytes(ddUpdated.Metadata.JSON, "vehicle_info.fuel_tank_capacity_gal").Float())

	// validate vin number created
	vinNumber, err := models.VinNumbers().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	assert.Equal(s.T(), vinNumber.Vin, vin)

}

// using existing WMI

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_CreatesDD() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021
	const wmi = "1FM"

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateWMI(s.T(), wmi, dm.ID, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &gateways.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2021",
		Make:                dm.Name,
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
	//s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(vinInfoResp, nil)

	deviceTypeInfo := make(map[string]interface{})
	deviceTypeInfo["mpg_city"] = vinInfoResp.MpgCity
	deviceTypeInfo["mpg_highway"] = vinInfoResp.MpgHighway
	deviceTypeInfo["mpg"] = vinInfoResp.Mpg
	deviceTypeInfo["base_msrp"] = vinInfoResp.MsrpBase
	deviceTypeInfo["fuel_tank_capacity_gal"] = vinInfoResp.FuelTankCapacityGal
	deviceTypeInfo["fuel_type"] = vinInfoResp.Fuel
	deviceTypeInfo["wheelbase"] = vinInfoResp.Wheelbase
	deviceTypeInfo["generation"] = vinInfoResp.Generation
	deviceTypeInfo["number_of_doors"] = vinInfoResp.Doors
	deviceTypeInfo["manufacturer_code"] = vinInfoResp.ManufacturerCode
	deviceTypeInfo["driven_wheels"] = vinInfoResp.Drive

	raw, _ := json.Marshal(vinInfoResp)
	yr, _ := strconv.Atoi(vinInfoResp.Year)
	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: buildStyleName(vinInfoResp),
		SubModel:  vinInfoResp.SubModel,
		Source:    "drivly",
		Year:      int32(yr),
		Model:     vinInfoResp.Model,
		Raw:       raw,
	}

	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)

	iceValue := "ICE"
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&iceValue, nil)
	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	s.NotNil(result, "expected result not nil")

	ddCreated, err := models.DeviceDefinitions().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(ddCreated.ID, result.DeviceDefinitionId)
	s.Assert().Equal(dm.ID, result.DeviceMakeId)
	// validate style was created
	ds, err := models.DeviceStyles().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)
	s.Assert().Equal(vinInfoResp.Trim+" "+vinInfoResp.SubModel, ds.Name)
	s.Assert().Equal(vinInfoResp.SubModel, ds.SubModel)
	s.Assert().Equal("drivly", ds.Source)
	s.Assert().Equal(ds.ExternalStyleID, common.SlugString(vinInfoResp.Trim+" "+vinInfoResp.SubModel))
	// validate vin number was create
	vn, err := models.VinNumbers().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	s.Assert().True(vn.DrivlyData.Valid)
	s.Assert().Equal(2021, vn.Year)
	s.Assert().Equal(ddCreated.ID, vn.DeviceDefinitionID)
	s.Assert().Equal(wmi, vn.Wmi)
	s.Assert().Equal("drivly", vn.DecodeProvider.String)
	s.Assert().Equal(vin, vn.Vin)
	s.Assert().Equal(vinInfoResp.Model, gjson.GetBytes(vn.DrivlyData.JSON, "model").String())

	// validate metadata was set
	assert.Equal(s.T(), vinInfoResp.Wheelbase, gjson.GetBytes(ddCreated.Metadata.JSON, "vehicle_info.wheelbase").String())
	assert.Equal(s.T(), int64(vinInfoResp.Doors), gjson.GetBytes(ddCreated.Metadata.JSON, "vehicle_info.number_of_doors").Int())
	assert.Equal(s.T(), int64(vinInfoResp.MsrpBase), gjson.GetBytes(ddCreated.Metadata.JSON, "vehicle_info.base_msrp").Int())
	assert.Equal(s.T(), int64(vinInfoResp.Mpg), gjson.GetBytes(ddCreated.Metadata.JSON, "vehicle_info.mpg").Int())
	assert.Equal(s.T(), vinInfoResp.FuelTankCapacityGal, gjson.GetBytes(ddCreated.Metadata.JSON, "vehicle_info.fuel_tank_capacity_gal").Float())
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingDD_AndStyleAndMetadata() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &gateways.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2021",
		Make:                dm.Name,
		Model:               dd.Model,
		SubModel:            "Hybrid",
		Trim:                "XLE",
		Doors:               4,
		MsrpBase:            23000,
		Fuel:                "Hybrid",
		FuelTankCapacityGal: 15.5,
		Mpg:                 25,
		Wheelbase:           "106 WB",
	}

	deviceTypeInfo := make(map[string]interface{})
	deviceTypeInfo["mpg_city"] = vinInfoResp.MpgCity
	deviceTypeInfo["mpg_highway"] = vinInfoResp.MpgHighway
	deviceTypeInfo["mpg"] = vinInfoResp.Mpg
	deviceTypeInfo["base_msrp"] = vinInfoResp.MsrpBase
	deviceTypeInfo["fuel_tank_capacity_gal"] = vinInfoResp.FuelTankCapacityGal
	deviceTypeInfo["fuel_type"] = vinInfoResp.Fuel
	deviceTypeInfo["wheelbase"] = vinInfoResp.Wheelbase
	deviceTypeInfo["generation"] = vinInfoResp.Generation
	deviceTypeInfo["number_of_doors"] = vinInfoResp.Doors
	deviceTypeInfo["manufacturer_code"] = vinInfoResp.ManufacturerCode
	deviceTypeInfo["driven_wheels"] = vinInfoResp.Drive

	yr, _ := strconv.Atoi(vinInfoResp.Year)
	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: buildStyleName(vinInfoResp),
		SubModel:  vinInfoResp.SubModel,
		Source:    "drivly",
		Year:      int32(yr),
		Make:      dm.Name,
		Model:     dd.Model,
	}

	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)

	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	// db setup
	ds := dbtesthelper.SetupCreateStyle(s.T(), dd.ID, buildStyleName(vinInfoResp), "drivly", vinInfoResp.SubModel, s.pdb)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DeviceDefinitionId)
	s.Assert().Equal(dm.ID, result.DeviceMakeId)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)

	// validate metadata was not changed - currently we only support updating it if no vehicle_info, if there is data leave as is
	ddUpdated, err := models.DeviceDefinitions().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal("defaultValue", gjson.GetBytes(ddUpdated.Metadata.JSON, "vehicle_info.mpg").String())
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingWMI() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)
	wmi := models.Wmi{
		Wmi:          "1FM",
		DeviceMakeID: dm.ID,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &gateways.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2021",
		Make:                dm.Name,
		Model:               dd.Model,
		SubModel:            "Hybrid",
		Trim:                "XLE",
		Doors:               4,
		MsrpBase:            23000,
		Fuel:                "Hybrid",
		FuelTankCapacityGal: 15.5,
		Mpg:                 25,
		Wheelbase:           "106 WB",
	}

	deviceTypeInfo := make(map[string]interface{})
	deviceTypeInfo["mpg_city"] = vinInfoResp.MpgCity
	deviceTypeInfo["mpg_highway"] = vinInfoResp.MpgHighway
	deviceTypeInfo["mpg"] = vinInfoResp.Mpg
	deviceTypeInfo["base_msrp"] = vinInfoResp.MsrpBase
	deviceTypeInfo["fuel_tank_capacity_gal"] = vinInfoResp.FuelTankCapacityGal
	deviceTypeInfo["fuel_type"] = vinInfoResp.Fuel
	deviceTypeInfo["wheelbase"] = vinInfoResp.Wheelbase
	deviceTypeInfo["generation"] = vinInfoResp.Generation
	deviceTypeInfo["number_of_doors"] = vinInfoResp.Doors
	deviceTypeInfo["manufacturer_code"] = vinInfoResp.ManufacturerCode
	deviceTypeInfo["driven_wheels"] = vinInfoResp.Drive

	yr, _ := strconv.Atoi(vinInfoResp.Year)
	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: buildStyleName(vinInfoResp),
		SubModel:  vinInfoResp.SubModel,
		Source:    "drivly",
		Year:      int32(yr),
		Model:     vinInfoResp.Model,
	}

	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)

	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DeviceDefinitionId)
	s.Assert().Equal(dm.ID, result.DeviceMakeId)
	// validate same number of wmi's
	wmis, err := models.Wmis().All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Len(wmis, 1)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingVINNumber() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)
	wmi := models.Wmi{
		Wmi:          "1FM",
		DeviceMakeID: dm.ID,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	// insert into vin numbers
	v := shared.VIN(vin)
	vinNumb := models.VinNumber{
		Vin:                vin,
		Wmi:                v.Wmi(),
		VDS:                v.VDS(),
		CheckDigit:         v.CheckDigit(),
		SerialNumber:       v.SerialNumber(),
		Vis:                v.VIS(),
		DeviceMakeID:       dm.ID,
		DeviceDefinitionID: dd.ID,
		Year:               2021,
		DecodeProvider:     null.StringFrom("drivly"),
	}
	err = vinNumb.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DeviceDefinitionId)
	s.Assert().Equal(dm.ID, result.DeviceMakeId)
	// validate same number of wmi's
	wmis, err := models.Wmis().All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Len(wmis, 1)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_InvalidVINYear_Vincario() {
	const vin = "1FMCU0G61QUA52727" // invalid year digit 10 - Q
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source: "vincario",
		Year:   2017,
		Make:   dm.Name,
		Model:  "Escape",
	}
	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.VincarioProvider).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	assert.NotNil(s.T(), qryResult)
	assert.NoError(s.T(), err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	assert.Equal(s.T(), int32(2017), result.Year)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_InvalidStyleName_Vincario() {
	const vin = "1FMCU0G61QUA52727" // invalid year digit 10 - Q
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source:    "vincario",
		Year:      2017,
		Make:      dm.Name,
		Model:     "Escape",
		StyleName: "1",
	}
	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.VincarioProvider).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	assert.NotNil(s.T(), qryResult)
	assert.NoError(s.T(), err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	assert.Equal(s.T(), int32(2017), result.Year)
	assert.Equal(s.T(), "", result.DeviceStyleId)

	count, err := models.VinNumbers().Count(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), count, "expected a new vin number to be inserted")
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Fail_ErrDecodeProvider_PartialDecode() {
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	wmi := models.Wmi{
		Wmi:          "1FM",
		DeviceMakeID: dm.ID,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source: "drivly",
	}
	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, errors.New("could not decode"))

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	assert.NotNil(s.T(), qryResult)
	assert.Error(s.T(), err, "failed to decode vin")
	// partial decode
	result := qryResult.(*p_grpc.DecodeVinResponse)
	s.Assert().Equal(int32(2021), result.Year)
	//s.Assert().Equal(dm.ID, result.DeviceMakeId)
	// future - another test for decode model when we have the info
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Fail_DecodeProviderBlankModel() {
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source: "vincario",
		Model:  "",
		Make:   "Ford",
	}
	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin})
	assert.Nil(s.T(), qryResult)
	assert.Error(s.T(), err, "decoded model name is blank")
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_DecodeKnownFallback() {
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	_ = dbtesthelper.SetupCreateWMI(s.T(), "1FM", dm.ID, s.pdb)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source: "vincario",
		Model:  "",
		Make:   "Ford",
	}
	s.mockVINService.EXPECT().GetVIN(vin, gomock.Any(), coremodels.AllProviders).Times(1).Return(vinDecodingInfoData, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin,
		KnownYear:  2022,
		KnownModel: "Bronco"})
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), qryResult)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	assert.Equal(s.T(), int32(2022), result.Year)
	assert.Equal(s.T(), dm.ID, result.DeviceMakeId)
}

func buildStyleName(vinInfo *gateways.DrivlyVINResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}
