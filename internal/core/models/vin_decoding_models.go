//nolint:tagliatelle
package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/volatiletech/null/v8"
)

type DecodeProviderEnum string

const (
	DrivlyProvider   DecodeProviderEnum = "drivly"
	VincarioProvider DecodeProviderEnum = "vincario"
	AutoIsoProvider  DecodeProviderEnum = "autoiso"
	DATGroupProvider DecodeProviderEnum = "dat"
	AllProviders     DecodeProviderEnum = ""
	TeslaProvider    DecodeProviderEnum = "tesla"
)

type VINDecodingInfoData struct {
	VIN        string
	Make       string
	Model      string
	SubModel   string
	Year       int32
	StyleName  string
	Source     DecodeProviderEnum
	ExternalID string
	MetaData   null.JSON
	Raw        []byte
	FuelType   string
}

type AutoIsoVINResponse struct {
	Version          string `json:"version"`
	Vin              string `json:"vin"`
	APIStatus        string `json:"apiStatus"`
	ResponseDate     string `json:"responseDate"`
	FunctionName     string `json:"functionName"`
	FunctionResponse struct {
		Data struct {
			API struct {
				CoreVersion     string `json:"core_version"`
				EndpointVersion int    `json:"endpoint_version"`
				JSONVersion     string `json:"json_version"`
				APIType         string `json:"api_type"`
				APICache        string `json:"api_cache"`
				DataPrecision   int    `json:"data_precision"`
				DataMatching    string `json:"data_matching"`
				LexLang         string `json:"lex_lang"`
			} `json:"api"`
			Analyze struct {
				VinOrginal struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_orginal"`
				VinCorrected struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_corrected"`
				VinYear struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_year"`
				VinSerial struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_serial"`
				Checkdigit struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"checkdigit"`
			} `json:"analyze"`
			Decoder struct {
				Make struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"make"`
				Model struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"model"`
				ModelYear struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"model_year"`
				Body struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"body"`
				FuelType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"fuel_type"`
				VehicleType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vehicle_type"`
				Doors struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"doors"`
				EngineDisplCm3 struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_displ_cm3"`
				EngineDisplL struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_displ_l"`
				EnginePowerHp struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_power_hp"`
				EnginePowerKw struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_power_kw"`
				EngineConf struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_conf"`
				EngineType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_type"`
				EngineVersion struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_version"`
				EngineHead struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_head"`
				EngineValves struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_valves"`
				EngineCylinders struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_cylinders"`
				EngineDisplCid struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_displ_cid"`
				EngineTurbo struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_turbo"`
				DriveType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"drive_type"`
				GearboxType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"gearbox_type"`
				EmissionStd struct {
					Desc   string `json:"desc"`
					Value  string `json:"value"`
					Co2Gkm string `json:"co2_gkm"`
				} `json:"emission_std"`
			} `json:"decoder"`
		} `json:"data"`
	} `json:"functionResponse"`
	LicenseInfo struct {
		LicenseNumber         string `json:"licenseNumber"`
		ValidTo               string `json:"validTo"`
		RemainingCredits      int    `json:"remainingCredits"`
		RemainingMonthlyLimit int    `json:"remainingMonthlyLimit"`
		RemainingDailyLimit   int    `json:"remainingDailyLimit"`
	} `json:"licenseInfo"`
}

// GetStyle returns a standard style string built from the data we have
func (v *AutoIsoVINResponse) GetStyle() string {
	s := ""
	// eg. Diesel
	if len(v.FunctionResponse.Data.Decoder.FuelType.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.FuelType.Value + " "
	}
	// eg. TDI
	if len(v.FunctionResponse.Data.Decoder.EngineVersion.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.EngineVersion.Value + " "
	}
	// eg. Manual
	if len(v.FunctionResponse.Data.Decoder.GearboxType.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.GearboxType.Value + " "
	}
	// eg. FWD / AWD
	if len(v.FunctionResponse.Data.Decoder.DriveType.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.DriveType.Value + " "
	}

	return strings.TrimSpace(s)
}

// GetSubModel returns the Body type, which we can use as the sub model.
func (v *AutoIsoVINResponse) GetSubModel() string {
	return strings.TrimSpace(v.FunctionResponse.Data.Decoder.Body.Value)
}

type VincarioInfoResponse struct {
	VIN                      string   `key:"VIN"`
	VehicleID                int      `key:"Vehicle ID"`
	Make                     string   `key:"Make"`
	Model                    string   `key:"Model"`
	ModelYear                int      `key:"Model Year"`
	ProductType              string   `key:"Product Type"`
	Body                     string   `key:"Body"`
	Series                   string   `key:"Series"`
	Drive                    string   `key:"Drive"`
	EngineDisplacement       int      `key:"Engine Displacement (ccm)"`
	FuelType                 string   `key:"Fuel Type - Primary"`
	EngineCode               string   `key:"Engine Code"`
	Transmission             string   `key:"Transmission"`
	NumberOfGears            int      `key:"Number of Gears"`
	EmissionStandard         string   `key:"Emission Standard"`
	PlantCountry             string   `key:"Plant Country"`
	ProductionStopped        int      `key:"Production Stopped"`
	EngineManufacturer       string   `key:"Engine Manufacturer"`
	EngineType               string   `key:"Engine Type"`
	AverageCO2Emission       float64  `key:"Average CO2 Emission (g/km)"`
	NumberOfWheels           int      `key:"Number Wheels"`
	NumberOfAxles            int      `key:"Number of Axles"`
	NumberOfDoors            int      `key:"Number of Doors"`
	FrontBrakes              string   `key:"Front Brakes"`
	RearBrakes               string   `key:"Rear Brakes"`
	BrakeSystem              string   `key:"Brake System"`
	SteeringType             string   `key:"Steering Type"`
	WheelSize                string   `key:"Wheel Size"`
	WheelSizeArray           []string `key:"Wheel Size Array"`
	Wheelbase                int      `key:"Wheelbase (mm)"`
	WheelbaseArray           []int    `key:"Wheelbase Array (mm)"`
	Height                   int      `key:"Height (mm)"`
	Length                   int      `key:"Length (mm)"`
	Width                    int      `key:"Width (mm)"`
	RearOverhang             int      `key:"Rear Overhang (mm)"`
	FrontOverhang            int      `key:"Front Overhang (mm)"`
	TrackFront               int      `key:"Track Front (mm)"`
	TrackRear                int      `key:"Track Rear (mm)"`
	MaxSpeed                 int      `key:"Max Speed (km/h)"`
	WeightEmpty              int      `key:"Weight Empty (kg)"`
	MaxWeight                int      `key:"Max Weight (kg)"`
	MaxRoofLoad              int      `key:"Max roof load (kg)"`
	TrailerLoadWithoutBrakes int      `key:"Permitted trailer load without brakes (kg)"`
	CheckDigit               string   `key:"Check Digit"`
	SequentialNumber         string   `key:"Sequential Number"`
}

// GetStyle returns a standard style string built from the data we have
func (v *VincarioInfoResponse) GetStyle() string {
	s := ""
	if len(v.FuelType) > 0 {
		s += v.FuelType + " "
	}
	if len(v.EngineType) > 0 {
		s += v.EngineType + " "
	}
	if len(v.Transmission) > 0 {
		s += v.Transmission + " "
	}
	if v.NumberOfGears > 0 {
		s += fmt.Sprintf("%d-speed", v.NumberOfGears)
	}

	return strings.TrimSpace(s)
}

// GetSubModel returns the Body type from Vincario, which we can use as the sub model.
func (v *VincarioInfoResponse) GetSubModel() string {
	return strings.TrimSpace(v.Body)
}

// GetMetadata returns a map of metadata for the vehicle, in standard format.
func (v *VincarioInfoResponse) GetMetadata() (null.JSON, error) {

	metadata := map[string]interface{}{
		"fuel_type":              v.FuelType,
		"driven_wheels":          v.Drive,
		"number_of_doors":        v.NumberOfDoors,
		"base_msrp":              nil,
		"epa_class":              v.EmissionStandard,
		"vehicle_type":           v.Body,
		"mpg_highway":            nil,
		"mpg_city":               nil,
		"fuel_tank_capacity_gal": nil,
		"mpg":                    nil,
	}

	bytes, err := json.Marshal(metadata)

	if err != nil {
		return null.JSON{}, err
	}

	return null.JSONFrom(bytes), nil
}

type DrivlyVINResponse struct {
	Vin                      string   `json:"vin"`
	WindowSticker            string   `json:"windowSticker"`
	Year                     string   `json:"year"`
	Make                     string   `json:"make"`
	Model                    string   `json:"model"`
	SubModel                 string   `json:"subModel"`
	Trim                     string   `json:"trim"`
	Generation               int      `json:"generation"`
	SubGeneration            int      `json:"subGeneration"`
	ManufacturerCode         string   `json:"manufacturerCode"`
	Body                     string   `json:"body"`
	Style                    string   `json:"style"`
	Type                     string   `json:"type"`
	Drive                    string   `json:"drive"`
	Transmission             string   `json:"transmission"`
	TransmissionDetails      string   `json:"transmissionDetails"`
	Engine                   string   `json:"engine"`
	EngineDetails            string   `json:"engineDetails"`
	Doors                    int      `json:"doors"`
	PaintColor               string   `json:"paintColor"`
	PaintName                string   `json:"paintName"`
	PaintCode                string   `json:"paintCode"`
	Interior                 string   `json:"interior"`
	Options                  []string `json:"options"`
	OptionCodes              string   `json:"optionCodes"`
	MsrpBase                 float64  `json:"msrpBase"`
	MsrpDiscount             float64  `json:"msrpDiscount"`
	MsrpOptions              float64  `json:"msrpOptions"`
	MsrpDelivery             float64  `json:"msrpDelivery"`
	Msrp                     float64  `json:"msrp"`
	WarrantyBasicMonths      int      `json:"warrantyBasicMonths"`
	WarrantyCorrosionMonths  int      `json:"warrantyCorrosionMonths"`
	WarrantyEmissionsMonths  int      `json:"warrantyEmissionsMonths"`
	WarrantyFullMonths       int      `json:"warrantyFullMonths"`
	WarrantyFullMiles        int      `json:"warrantyFullMiles"`
	WarrantyDrivetrainMonths int      `json:"warrantyDrivetrainMonths"`
	WarrantyPowertrainMonths int      `json:"warrantyPowertrainMonths"`
	WarrantyPowertrainMiles  int      `json:"warrantyPowertrainMiles"`
	WarrantyRoadsideMonths   int      `json:"warrantyRoadsideMonths"`
	WarrantyRoadsideMiles    int      `json:"warrantyRoadsideMiles"`
	Wheelbase                string   `json:"wheelbase"`
	Fuel                     string   `json:"fuel"`
	FuelTankCapacityGal      float64  `json:"fuelTankCapacityGal"`
	Mpg                      int      `json:"mpg"`
	MpgCity                  int      `json:"mpgCity"`
	MpgHighway               int      `json:"mpgHighway"`
	LastOdometer             int      `json:"lastOdometer"`
	LastOdometerDate         string   `json:"lastOdometerDate"`
	EstimatedOdometer        int      `json:"estimatedOdometer"`
	Salvage                  bool     `json:"salvage"`
	PreviousOwners           int      `json:"previousOwners"`
	TotalLoss                bool     `json:"totalLoss"`
	Branded                  bool     `json:"branded"`
	LastTitleState           string   `json:"lastTitleState"`
	TitleIssueDate           string   `json:"titleIssueDate"`
	TitleNumber              string   `json:"titleNumber"`
	Confidence               float64  `json:"confidence"`
	VehicleHistory           []string `json:"vehicleHistory"`
	InstalledEquipment       []string `json:"installedEquipment"`
	Dimensions               []string `json:"dimensions"`
}

// GetExternalID builds something we can use as an external ID that is drivly specific, at the MMY level (not for style)
func (vir *DrivlyVINResponse) GetExternalID() string {
	// cant use shared.SlugString due to import cycle
	return strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s-%s-%s", vir.Make, vir.Model, vir.Year)), " ", "")
}

type DATGroupInfoResponse struct {
	VIN              string `json:"vin"`
	DatECode         string `json:"datecode"`
	SalesDescription string `json:"salesDescription"`
	VehicleTypeName  string `json:"vehicleTypeName"`
	// make
	ManufacturerName string `json:"manufacturerName"`
	BaseModelName    string `json:"baseModelName"`
	SubModelName     string `json:"subModelName"`
	// this is the model name we want to use
	MainTypeGroupName string `json:"mainTypeGroupName"`
	VinAccuracy       int    `json:"vinAccuracy"`

	// when we're unable to get exact year
	YearLow  int `json:"yearLow"`
	YearHigh int `json:"yearHigh"`
	// we don't always get the exact year
	Year int `json:"year"`

	SeriesEquipment   []DATGroupEquipment `json:"seriesEquipment"`
	SpecialEquipment  []DATGroupEquipment `json:"specialEquipment"`
	DATECodeEquipment []DATGroupEquipment `json:"datECodeEquipment"`
	VINEquipment      []DATGroupEquipment `json:"vinEquipments"`
}

// nolint
type DATGroupEquipment struct {
	DatEquipmentId          string `json:"datEquipmentId"`
	ManufacturerEquipmentId string `json:"manufacturerEquipmentId"`
	// if Vin Equipment, this comes from ShortName
	ManufacturerDescription string `json:"manufacturerDescription"`
	Description             string `json:"description"`
}

// nolint
type Japan17VINResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ModelComparisonInAutohome string `json:"model_comparison_in_autohome"`
		FullVin                   string `json:"full_vin"`
		ModelYearFromVin          string `json:"model_year_from_vin"`
		Epc                       string `json:"epc"`
		EpcCn                     string `json:"epc_cn"`
		MyModelStdId              string `json:"my_model_std_id"`
		EpcId                     string `json:"epc_id"`
		GroupId                   string `json:"group_id"`
		MatchingMode              string `json:"matching_mode"`
		Brand                     string `json:"brand"`
		GonggaoNo                 string `json:"gonggao_no"`
		GonggaoNoMatchingType     string `json:"gonggao_no_matching_type"`
		MadeInCn                  string `json:"made_in_cn"`
		MadeInEn                  string `json:"made_in_en"`
		BuildDate                 string `json:"build_date"`
		ModelList                 []struct {
			Id                   int    `json:"Id"`
			JsId                 int    `json:"Js_id"`
			ModelDetailKey       string `json:"Model_detail_key"`
			GonggaoNo            string `json:"Gonggao_no"`
			GroupId              string `json:"Group_id"`
			UrlMake              string `json:"UrlMake"`
			Epc                  string `json:"Epc"`
			EpcId                string `json:"Epc_id"`
			ChassisCode          string `json:"Chassis_code"`
			ModelYear            string `json:"Model_year"`
			ModelDetail          string `json:"Model_detail"`
			ModelDetailEn        string `json:"Model_detail_en"`
			Factory              string `json:"Factory"`
			FactoryEn            string `json:"Factory_en"`
			Brand                string `json:"Brand"`
			BrandEn              string `json:"Brand_en"`
			Series               string `json:"Series"`
			SeriesEn             string `json:"Series_en"`
			Model                string `json:"Model"`
			ModelEn              string `json:"Model_en"`
			SalesVersion         string `json:"Sales_version"`
			SalesVersionEn       string `json:"Sales_version_en"`
			Cc                   string `json:"Cc"`
			CcEn                 string `json:"Cc_en"`
			EngineNo             string `json:"Engine_no"`
			EngineNoEn           string `json:"Engine_no_en"`
			Kw                   string `json:"Kw"`
			Hp                   string `json:"Hp"`
			AirIntake            string `json:"Air_intake"`
			AirIntakeEn          string `json:"Air_intake_en"`
			FuelType             string `json:"Fuel_type"`
			FuelTypeEn           string `json:"Fuel_type_en"`
			EffluentStandard     string `json:"Effluent_standard"`
			EffluentStandardEn   string `json:"Effluent_standard_en"`
			TransmissionDetail   string `json:"Transmission_detail"`
			TransmissionDetailEn string `json:"Transmission_detail_en"`
			GearNum              string `json:"Gear_num"`
			GearNumEn            string `json:"Gear_num_en"`
			TransCode            string `json:"Trans_code"`
			DrivingMode          string `json:"Driving_mode"`
			DrivingModeEn        string `json:"Driving_mode_en"`
			DoorNum              string `json:"Door_num"`
			DoorNumEn            string `json:"Door_num_en"`
			SeatNum              string `json:"Seat_num"`
			BodyType             string `json:"Body_type"`
			BodyTypeEn           string `json:"Body_type_en"`
			DateBegin            string `json:"Date_begin"`
			DateEnd              string `json:"Date_end"`
			Price                string `json:"Price"`
			PriceUnit            string `json:"Price_unit"`
			AutohomeId           string `json:"Autohome_id"`
			ImgAdress            string `json:"Img_adress"`
			XsId                 string `json:"Xs_id"`
			SalesName            string `json:"Sales_name"`
			SalesNameEn          string `json:"Sales_name_en"`
			SeriesZh             string `json:"Series_zh"`
			ModelZh              string `json:"Model_zh"`
		} `json:"model_list"`
		ModelOriginalEpcList []struct {
			EpcId         int `json:"Epc_id"`
			CarAttributes []struct {
				Language         string `json:"Language"`
				IsMajorAttribute bool   `json:"IsMajorAttribute"`
				ColName          string `json:"Col_name"`
				ColValue         string `json:"Col_value"`
			} `json:"CarAttributes"`
		} `json:"model_original_epc_list"`
		ModelGonggaoList interface{} `json:"model_gonggao_list"`
		ModelImportList  interface{} `json:"model_import_list"`
	} `json:"data"`
}
