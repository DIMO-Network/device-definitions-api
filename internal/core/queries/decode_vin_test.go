package queries

import (
	"context"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
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

	ctrl             *gomock.Controller
	pdb              db.Store
	container        testcontainers.Container
	ctx              context.Context
	mockDrivlyAPISvc *mock_gateways.MockDrivlyAPIService

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
	repo := repositories.NewDeviceDefinitionRepository(s.pdb.DBS)
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)
	s.queryHandler = NewDecodeVINQueryHandler(s.pdb.DBS, s.mockDrivlyAPISvc, repo, dbtesthelper.Logger())
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
	vinInfoResp := &gateways.VINInfoResponse{
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
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(vinInfoResp, nil)
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
}

// using existing WMI

func (s *DecodeVINQueryHandlerSuite) TestHandle_Success_CreatesDD() {
	const vin = "1FMCU0G61MUA52727" // ford escape 2021
	const wmi = "1FM"

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	_ = dbtesthelper.SetupCreateAutoPiIntegration(s.T(), s.pdb)
	_ = dbtesthelper.SetupCreateWMI(s.T(), wmi, dm.ID, s.pdb)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &gateways.VINInfoResponse{
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
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(vinInfoResp, nil)

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
	vinInfoResp := &gateways.VINInfoResponse{
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
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(vinInfoResp, nil)
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

	dm := dbtesthelper.SetupCreateMake(s.T(), "Ford", s.pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinitionWithVehicleInfo(s.T(), dm, "Escape", 2021, s.pdb)
	wmi := models.Wmi{
		Wmi:          "1FM",
		DeviceMakeID: dm.ID,
	}
	err := wmi.Insert(s.ctx, s.pdb.DBS().Writer, boil.Infer())
	s.Require().NoError(err)

	// mock setup, include some attributes we should expect in metadata, and trim we should expect created in styles
	vinInfoResp := &gateways.VINInfoResponse{
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
	s.mockDrivlyAPISvc.EXPECT().GetVINInfo(vin).Times(1).Return(vinInfoResp, nil)

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

func Test_drivlyVINInfoToUpdateAttr(t *testing.T) {

	vinInfo := &gateways.VINInfoResponse{
		Wheelbase:           "106 WB",
		MsrpBase:            32123,
		Mpg:                 24,
		FuelTankCapacityGal: 23.25,
	}
	got := drivlyVINInfoToUpdateAttr(vinInfo)
	// need a lookup for that
	assert.Equal(t, "106 WB", lookupUpdateAttr(got, "wheelbase"))
	assert.Equal(t, "32123", lookupUpdateAttr(got, "base_msrp"))
	assert.Equal(t, "24", lookupUpdateAttr(got, "mpg"))
	assert.Equal(t, "23.25", lookupUpdateAttr(got, "fuel_tank_capacity_gal"))
}

func lookupUpdateAttr(items []*coremodels.UpdateDeviceTypeAttribute, name string) string {
	for _, item := range items {
		if item.Name == name {
			return item.Value
		}
	}
	return ""
}
