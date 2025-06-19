package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	mock_repository "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories/mocks"
	"github.com/segmentio/ksuid"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	"github.com/volatiletech/null/v8"

	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	vinutil "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

type DecodeVINQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl                               *gomock.Controller
	pdb                                db.Store
	container                          testcontainers.Container
	ctx                                context.Context
	mockVINService                     *mock_services.MockVINDecodingService
	mockFuelAPIService                 *mock_gateways.MockFuelAPIService
	mockPowerTrainTypeService          *mock_services.MockPowerTrainTypeService
	mockDeviceDefinitionOnChainService *mock_gateways.MockDeviceDefinitionOnChainService

	queryHandler DecodeVINQueryHandler
	mockVINRepo  *mock_repository.MockVINRepository
	mockIdentity *mock_gateways.MockIdentityAPI
}

const country = "USA"

func TestDecodeVINQueryHandler(t *testing.T) {
	suite.Run(t, new(DecodeVINQueryHandlerSuite))
}

func (s *DecodeVINQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.ctx = context.Background()

	s.mockVINService = mock_services.NewMockVINDecodingService(s.ctrl)
	s.mockPowerTrainTypeService = mock_services.NewMockPowerTrainTypeService(s.ctrl)
	s.mockDeviceDefinitionOnChainService = mock_gateways.NewMockDeviceDefinitionOnChainService(s.ctrl)
	s.mockFuelAPIService = mock_gateways.NewMockFuelAPIService(s.ctrl)

	s.mockVINRepo = mock_repository.NewMockVINRepository(s.ctrl)
	s.mockIdentity = mock_gateways.NewMockIdentityAPI(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)
	s.queryHandler = NewDecodeVINQueryHandler(s.pdb.DBS, s.mockVINService, s.mockVINRepo, dbtesthelper.Logger(), s.mockFuelAPIService, s.mockPowerTrainTypeService, s.mockDeviceDefinitionOnChainService, s.mockIdentity)
}

func (s *DecodeVINQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *DecodeVINQueryHandlerSuite) TearDownSuite() {
	fmt.Printf("shutting down postgres at with session: %s \n", s.container.SessionID())
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingDD_UpdatesAttributes_CreatesStyle() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	dm := dbtesthelper.SetupCreateMake("Ford")
	//s.mockIdentity.EXPECT().GetManufacturer("ford").Return(&dm, nil)
	dd := dbtesthelper.SetupCreateDeviceDefinition(s.T(), dm.Name, "Escape", 2021, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &coremodels.DrivlyVINResponse{
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
	definitionID := dd.ID
	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo(vinDecodingInfoData.StyleName, vinDecodingInfoData.FuelType).Return("ICE")
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, dd.Model, int(dd.Year)), nil, nil)
	wmiDb := &models.Wmi{
		Wmi:              vin[:3],
		ManufacturerName: dm.Name,
	}
	s.mockVINRepo.EXPECT().GetOrCreateWMI(gomock.Any(), vin[:3], dm.Name).Return(wmiDb, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DefinitionId)

	// validate style was created
	ds, err := models.DeviceStyles().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)
	s.Assert().Equal(vinInfoResp.Trim+" "+vinInfoResp.SubModel, ds.Name)
	s.Assert().Equal(vinInfoResp.SubModel, ds.SubModel)
	s.Assert().Equal("drivly", ds.Source)
	s.Assert().Equal(ds.ExternalStyleID, stringutils.SlugString(vinInfoResp.Trim+" "+vinInfoResp.SubModel))

	// validate metadata was updated on DD
	// validate vin number created
	vinNumber, err := models.VinNumbers().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	assert.Equal(s.T(), vinNumber.Vin, vin)

}

// WMI oem conflict, same WMI for different Make name is ok. Ford WMI already exists, but decodes to Lincoln w/ same WMI
func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_CreatesDD_WithMismatchWMI() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // Lincoln escape 2021
	const wmi = "1FM"

	dmFord := dbtesthelper.SetupCreateMake("Ford")
	//s.mockIdentity.EXPECT().GetManufacturer("ford").Return(&dmFord, nil)
	dmLincoln := dbtesthelper.SetupCreateMake("Lincoln")
	//s.mockIdentity.EXPECT().GetManufacturer("lincoln").Return(&dmLincoln, nil)
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateWMI(s.T(), wmi, dmFord.Name, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &coremodels.DrivlyVINResponse{
		Vin:                 vin,
		Year:                "2022",
		Make:                dmLincoln.Name,
		Model:               "Aviator",
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

	raw, _ := json.Marshal(vinInfoResp)
	yr, _ := strconv.Atoi(vinInfoResp.Year)
	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		StyleName: buildStyleName(vinInfoResp),
		SubModel:  vinInfoResp.SubModel,
		Make:      vinInfoResp.Make,
		Source:    "drivly",
		Year:      int32(yr),
		Model:     vinInfoResp.Model,
		Raw:       raw,
	}
	definitionID := "lincoln_aviator_2022"
	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)

	styleLevelPT := "PHEV"
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		nil, nil, nil) // should return nil b/c doesn't exist
	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo(vinDecodingInfoData.StyleName, vinDecodingInfoData.FuelType).Return(styleLevelPT)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, dmLincoln.Name, gomock.Any()).Return(&trxHashHex, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(vinInfoResp.Make, vinInfoResp.Model, 2022, gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	s.NotNil(result, "expected result not nil")

	// validate style was created
	ds, err := models.DeviceStyles().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)
	s.Assert().Equal(vinInfoResp.Trim+" "+vinInfoResp.SubModel, ds.Name)
	s.Assert().Equal(vinInfoResp.SubModel, ds.SubModel)
	s.Assert().Equal("drivly", ds.Source)
	s.Assert().Equal(ds.ExternalStyleID, stringutils.SlugString(vinInfoResp.Trim+" "+vinInfoResp.SubModel))
	s.Assert().Equal(styleLevelPT, gjson.GetBytes(ds.Metadata.JSON, common.PowerTrainType).Str)
	// validate vin number was create
	vn, err := models.VinNumbers().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	s.Assert().True(vn.DrivlyData.Valid)
	s.Assert().Equal(2022, vn.Year)
	s.Assert().Equal(definitionID, vn.DefinitionID)
	s.Assert().Equal(wmi, vn.Wmi.String)
	s.Assert().Equal("drivly", vn.DecodeProvider.String)
	s.Assert().Equal(vin, vn.Vin)
	s.Assert().Equal(dmLincoln.Name, vn.ManufacturerName)
	s.Assert().Equal(vinInfoResp.Model, gjson.GetBytes(vn.DrivlyData.JSON, "model").String())

	// validate images was created
	ddImages, err := models.Images(models.ImageWhere.DefinitionID.EQ(definitionID)).All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().NotEmpty(ddImages)
}

// Japan
func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_JapanChassisNumber_existingVIN() {
	const vin = "ZWR90-8000186" // toyota something or other

	dm := dbtesthelper.SetupCreateMake("Toyota")
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Yaris", 2024, s.pdb)

	vinNumb := models.VinNumber{
		Vin:              vin,
		SerialNumber:     "8000186",
		ManufacturerName: dm.Name,
		DefinitionID:     dd.ID,
		Year:             2024,
		DecodeProvider:   null.StringFrom("manual"),
	}
	err := vinNumb.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	// mock setup for powertrain lookup, which is in the vin decode response
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), dd.ID).Return(
		&coremodels.DeviceDefinitionTablelandModel{
			ID:         dd.ID,
			Model:      dd.Model,
			Year:       dd.Year,
			DeviceType: common.DefaultDeviceType,
			Metadata: &coremodels.DeviceDefinitionMetadata{DeviceAttributes: []coremodels.DeviceTypeAttribute{
				{Name: "powertrain_type", Value: "ICE"},
			}},
		}, big.NewInt(1), nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2024), result.Year)
	s.Assert().Equal(dd.ID, result.DefinitionId)
	s.Assert().Equal(dm.Name, result.Manufacturer)
}

// using existing WMI
func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_CreatesDD() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // ford escape 2021
	const wmi = "1FM"

	dm := dbtesthelper.SetupCreateMake("Ford")
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateWMI(s.T(), wmi, dm.Name, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &coremodels.DrivlyVINResponse{
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
		Make:      vinInfoResp.Make,
		Source:    "drivly",
		Year:      int32(yr),
		Model:     vinInfoResp.Model,
		Raw:       raw,
	}
	definitionID := "ford_escape_2021"
	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)

	styleLevelPT := "PHEV"
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		nil, nil, nil) // should return nil b/c doesn't exist
	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo(vinDecodingInfoData.StyleName, vinDecodingInfoData.FuelType).Return(styleLevelPT)

	trxHashHex := "0xa90868fe9364dbf41695b3b87e630f6455cfd63a4711f56b64f631b828c02b35"
	s.mockDeviceDefinitionOnChainService.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(&trxHashHex, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages("Ford", "Escape", 2021, gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	s.NotNil(result, "expected result not nil")

	// validate style was created
	ds, err := models.DeviceStyles().One(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)
	s.Assert().Equal(vinInfoResp.Trim+" "+vinInfoResp.SubModel, ds.Name)
	s.Assert().Equal(vinInfoResp.SubModel, ds.SubModel)
	s.Assert().Equal("drivly", ds.Source)
	s.Assert().Equal(ds.ExternalStyleID, stringutils.SlugString(vinInfoResp.Trim+" "+vinInfoResp.SubModel))
	s.Assert().Equal(styleLevelPT, gjson.GetBytes(ds.Metadata.JSON, common.PowerTrainType).Str)
	// validate vin number was create
	vn, err := models.VinNumbers().One(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	s.Assert().True(vn.DrivlyData.Valid)
	s.Assert().Equal(2021, vn.Year)
	s.Assert().Equal(definitionID, vn.DefinitionID)
	s.Assert().Equal(wmi, vn.Wmi.String)
	s.Assert().Equal("drivly", vn.DecodeProvider.String)
	s.Assert().Equal(vin, vn.Vin)
	s.Assert().Equal(vinInfoResp.Model, gjson.GetBytes(vn.DrivlyData.JSON, "model").String())

	// validate images was created
	ddImages, err := models.Images(models.ImageWhere.DefinitionID.EQ(definitionID)).All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().NotEmpty(ddImages)
}

func buildTestTblDD(definitionID, model string, year int) *coremodels.DeviceDefinitionTablelandModel {
	return &coremodels.DeviceDefinitionTablelandModel{
		ID:         definitionID,
		KSUID:      ksuid.New().String(),
		Model:      model,
		Year:       year,
		DeviceType: "vehicle",
		ImageURI:   "",
		Metadata: &coremodels.DeviceDefinitionMetadata{DeviceAttributes: []coremodels.DeviceTypeAttribute{
			{Name: "powertrain_type", Value: "ICE"},
		}},
	}
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingDD_AndStyleAndMetadata() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	dm := dbtesthelper.SetupCreateMake("Ford")
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &coremodels.DrivlyVINResponse{
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
	definitionID := dd.ID

	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo(vinDecodingInfoData.StyleName, vinDecodingInfoData.FuelType).Return("HEV")
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, dd.Model, int(dd.Year)), nil, nil)
	// db setup
	ds := dbtesthelper.SetupCreateStyle(s.T(), definitionID, buildStyleName(vinInfoResp), "drivly", vinInfoResp.SubModel, s.pdb)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)
	wmiDb := &models.Wmi{
		Wmi:              vin[:3],
		ManufacturerName: dm.Name,
	}
	s.mockVINRepo.EXPECT().GetOrCreateWMI(gomock.Any(), vin[:3], dm.Name).Return(wmiDb, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DefinitionId)
	s.Assert().Equal(dm.Name, result.Manufacturer)
	s.Assert().Equal(ds.ID, result.DeviceStyleId)

}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingWMI() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake("Ford")
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)
	wmi := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: dm.Name,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &coremodels.DrivlyVINResponse{
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
		Make:      vinInfoResp.Make,
		Source:    "drivly",
		Year:      int32(yr),
		Model:     vinInfoResp.Model,
	}

	metaDataInfo := make(map[string]interface{})
	metaDataInfo["vehicle_info"] = deviceTypeInfo
	metaData, _ := json.Marshal(metaDataInfo)
	vinDecodingInfoData.MetaData = null.JSONFrom(metaData)
	definitionID := dd.ID

	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo(vinDecodingInfoData.StyleName, vinDecodingInfoData.FuelType).Return("HEV")
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, dd.Model, int(dd.Year)), nil, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DefinitionId)
	s.Assert().Equal(dm.Name, result.Manufacturer)
	// validate same number of wmi's
	wmis, err := models.Wmis().All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Len(wmis, 1)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_TeslaDecode() {
	ctx := context.Background()
	const vin = "5YJ3E1EA2PF696023" // tesla model 3 2023

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake("Tesla")
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Model 3", 2023, s.pdb)
	wmi := models.Wmi{
		Wmi:              "5YJ",
		ManufacturerName: dm.Name,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Make:   "Tesla",
		Source: "tesla",
		Year:   int32(2023),
		Model:  "Model 3",
	}

	definitionID := dd.ID

	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.TeslaProvider, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo(vinDecodingInfoData.StyleName, vinDecodingInfoData.FuelType).Return("BEV")
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, dd.Model, dd.Year), nil, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2023), result.Year)
	s.Assert().Equal(dd.ID, result.DefinitionId)
	s.Assert().Equal(dm.Name, result.Manufacturer)
	// validate same number of wmi's
	wmis, err := models.Wmis().All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Len(wmis, 1)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_WithExistingVINNumber() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake("Ford")
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)
	wmi := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: dm.Name,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	// insert into vin numbers
	v := vinutil.VIN(vin)
	vinNumb := models.VinNumber{
		Vin:              vin,
		Wmi:              null.StringFrom(v.Wmi()),
		VDS:              null.StringFrom(v.VDS()),
		CheckDigit:       null.StringFrom(v.CheckDigit()),
		SerialNumber:     v.SerialNumber(),
		Vis:              null.StringFrom(v.VIS()),
		ManufacturerName: dm.Name,
		DefinitionID:     dd.ID,
		Year:             2021,
		DecodeProvider:   null.StringFrom("drivly"),
	}
	err = vinNumb.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)
	// mock needed for powertrain lookup
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), dd.ID).Return(
		&coremodels.DeviceDefinitionTablelandModel{
			ID:         dd.ID,
			Model:      dd.Model,
			Year:       dd.Year,
			DeviceType: common.DefaultDeviceType,
			Metadata: &coremodels.DeviceDefinitionMetadata{DeviceAttributes: []coremodels.DeviceTypeAttribute{
				{Name: "powertrain_type", Value: "ICE"},
			}},
		}, big.NewInt(1), nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	s.NoError(err)
	result := qryResult.(*p_grpc.DecodeVinResponse)

	s.NotNil(result, "expected result not nil")
	s.Assert().Equal(int32(2021), result.Year)
	s.Assert().Equal(dd.ID, result.DefinitionId)
	s.Assert().Equal(dm.Name, result.Manufacturer)
	// validate same number of wmi's
	wmis, err := models.Wmis().All(s.ctx, s.pdb.DBS().Reader)
	s.Require().NoError(err)
	s.Assert().Len(wmis, 1)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_InvalidVINYear_AutoIso() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake("Ford")

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source: "vincario",
		Year:   2017,
		Make:   dm.Name,
		Model:  "Escape",
	}
	definitionID := "ford_escape_2017"
	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo("", "").Return("ICE") // normally this would return ""
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, "Escape", 2021), nil, nil)
	wmiDb := &models.Wmi{
		Wmi:              vin[:3],
		ManufacturerName: dm.Name,
	}
	s.mockVINRepo.EXPECT().GetOrCreateWMI(gomock.Any(), vin[:3], dm.Name).Return(wmiDb, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	assert.NotNil(s.T(), qryResult)
	assert.NoError(s.T(), err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	assert.Equal(s.T(), int32(2017), result.Year)
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_InvalidStyleName_AutoIso() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake("Ford")

	vinDecodingInfoData := &coremodels.VINDecodingInfoData{
		Source:    "vincario",
		Year:      2017,
		Make:      dm.Name,
		Model:     "Escape",
		StyleName: "1",
	}
	definitionID := "ford_escape_2017"
	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(vinDecodingInfoData, nil)
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo("1", "").Return("ICE")
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, "Escape", 2017), nil, nil)
	wmiDb := &models.Wmi{
		Wmi:              vin[:3],
		ManufacturerName: dm.Name,
	}
	s.mockVINRepo.EXPECT().GetOrCreateWMI(gomock.Any(), vin[:3], dm.Name).Return(wmiDb, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	assert.NotNil(s.T(), qryResult)
	assert.NoError(s.T(), err)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	assert.Equal(s.T(), int32(2017), result.Year)
	assert.Equal(s.T(), "", result.DeviceStyleId)

	count, err := models.VinNumbers().Count(s.ctx, s.pdb.DBS().Reader)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), count, "expected a new vin number to be inserted")
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Fail_DecodeErr() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateMake("Ford")

	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(nil, fmt.Errorf("unable to decode"))

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country})
	assert.Nil(s.T(), qryResult)
	assert.Error(s.T(), err, "unable to decode")
}

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_DecodeKnownFallback() {
	ctx := context.Background()
	const vin = "1FMCU0G61MUA52727" // invalid year digit 10 - Q

	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	dm := dbtesthelper.SetupCreateMake("Ford")
	_ = dbtesthelper.SetupCreateWMI(s.T(), "1FM", dm.Name, s.pdb)

	definitionID := "ford_bronco_2022"
	s.mockVINService.EXPECT().GetVIN(ctx, vin, coremodels.AllProviders, "USA").Times(1).Return(nil, fmt.Errorf("unable to decode"))
	s.mockPowerTrainTypeService.EXPECT().ResolvePowerTrainFromVinInfo("", "").Return("ICE")

	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), definitionID).Return(
		buildTestTblDD(definitionID, "Bronco", 20222), nil, nil)

	image := gateways.FuelImage{
		SourceURL: "https://image",
	}
	fuelDeviceImagesMock := gateways.FuelDeviceImages{
		FuelAPIID: "1",
		Height:    1,
		Width:     1,
		Images:    []gateways.FuelImage{image},
	}
	s.mockFuelAPIService.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(fuelDeviceImagesMock, nil)

	qryResult, err := s.queryHandler.Handle(s.ctx, &DecodeVINQuery{VIN: vin, Country: country,
		KnownYear:  2022,
		KnownModel: "Bronco"})
	// make will be inferred by WMI
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), qryResult)
	result := qryResult.(*p_grpc.DecodeVinResponse)
	assert.Equal(s.T(), int32(2022), result.Year)
	assert.Equal(s.T(), dm.Name, result.Manufacturer)
	assert.Equal(s.T(), "probably smartcar", result.Source)
	assert.NotEmptyf(s.T(), result.DefinitionId, "dd expected")
	assert.Equal(s.T(), result.NewTrxHash, "")
}

func TestResolveMetadataFromInfo(t *testing.T) {
	testCases := []struct {
		name       string
		powertrain string
		vinInfo    *coremodels.VINDecodingInfoData
		expectedMD *coremodels.DeviceDefinitionMetadata
	}{
		{
			name:       "valid powertrain and vinInfo",
			powertrain: "BEV",
			vinInfo:    &coremodels.VINDecodingInfoData{StyleName: "Test Style"},
			expectedMD: &coremodels.DeviceDefinitionMetadata{
				DeviceAttributes: []coremodels.DeviceTypeAttribute{
					{Name: "powertrain_type", Value: "BEV"},
				},
			},
		},
		{
			name:       "empty powertrain",
			powertrain: "",
			vinInfo:    &coremodels.VINDecodingInfoData{StyleName: "Test Style"},
			expectedMD: &coremodels.DeviceDefinitionMetadata{
				DeviceAttributes: []coremodels.DeviceTypeAttribute{},
			},
		},
		{
			name:       "nil vinInfo",
			powertrain: "PHEV",
			vinInfo:    nil,
			expectedMD: &coremodels.DeviceDefinitionMetadata{
				DeviceAttributes: []coremodels.DeviceTypeAttribute{
					{Name: "powertrain_type", Value: "PHEV"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualMD := resolveMetadataFromInfo(tc.powertrain, tc.vinInfo)
			assert.Equal(t, tc.expectedMD, actualMD)
		})
	}
}

func buildStyleName(vinInfo *coremodels.DrivlyVINResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}

func (s *DecodeVINQueryHandlerSuite) TestDecodeVINQueryHandler_vinInfoFromKnown_multipleWMI() {
	vinny := "1FMCU0G61MUA52727"
	v := vinutil.VIN(vinny)
	// insert into wmi
	wmi1 := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: "Ford",
	}
	wmi2 := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: "Lincoln",
	}
	err := wmi1.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	err = wmi2.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	// mock call to get definition by id
	definitionID := "ford_escape_2020"

	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), gomock.AnyOf("lincoln_escape_2020", definitionID)).AnyTimes().Return(&coremodels.DeviceDefinitionTablelandModel{
		ID:         definitionID,
		KSUID:      ksuid.New().String(),
		Model:      "Escape",
		Year:       2020,
		DeviceType: common.DefaultDeviceType,
		ImageURI:   "",
		Metadata:   nil,
	}, nil, nil)
	//s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), "lincoln_escape_2020").AnyTimes().Return(nil, nil, fmt.Errorf("not found"))

	got, err := s.queryHandler.vinInfoFromKnown(v, "Escape", 2020)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "1FMCU0G61MUA52727", got.VIN)
	assert.Equal(s.T(), "Ford", got.Make)
	assert.Equal(s.T(), "Escape", got.Model)
	assert.Equal(s.T(), int32(2020), got.Year)
}

func (s *DecodeVINQueryHandlerSuite) TestDecodeVINQueryHandler_vinInfoFromKnown_singleWMI() {
	vinny := "1FMCU0G61MUA52727"
	v := vinutil.VIN(vinny)
	// insert into wmi
	wmi1 := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: "Ford",
	}

	err := wmi1.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)

	got, err := s.queryHandler.vinInfoFromKnown(v, "Escape", 2020)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "1FMCU0G61MUA52727", got.VIN)
	assert.Equal(s.T(), "Ford", got.Make)
	assert.Equal(s.T(), "Escape", got.Model)
	assert.Equal(s.T(), int32(2020), got.Year)
}

func (s *DecodeVINQueryHandlerSuite) TestDecodeVINQueryHandler_vinInfoFromKnown_multipleWMINoDDFound() {
	vinny := "1FMCU0G61MUA52727"
	v := vinutil.VIN(vinny)
	// insert into wmi
	wmi1 := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: "Ford",
	}
	wmi2 := models.Wmi{
		Wmi:              "1FM",
		ManufacturerName: "Lincoln",
	}
	err := wmi1.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	err = wmi2.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	require.NoError(s.T(), err)
	// mock call to get definition by id
	definitionID := "ford_escape_2020"
	s.mockDeviceDefinitionOnChainService.EXPECT().GetDefinitionByID(gomock.Any(), gomock.AnyOf("lincoln_escape_2020", definitionID)).Times(2).Return(nil, nil, fmt.Errorf("not found"))

	got, err := s.queryHandler.vinInfoFromKnown(v, "Escape", 2020)
	require.Error(s.T(), err, "vinInfoFromKnown: unable to determine the right OEM between Ford, Lincoln for WMI %s 1FM")
	require.Nil(s.T(), got)
}
