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
	Japan17VIN       DecodeProviderEnum = "japan17vin"
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
type Japan17MMY struct {
	VIN                   string `json:"vin"`
	ManufacturerName      string `json:"manufacturerName"`
	ManufacturerLowerCase string `json:"manufacturerLowerCase"`
	ModelName             string `json:"modelName"`
	Year                  int    `json:"year"`
}

// nolint
type CarVxResponse struct {
	Data []struct {
		Make            string `json:"make"`
		Model           string `json:"model"`
		Grade           string `json:"grade"`
		Body            string `json:"body"`
		Engine          string `json:"engine"`
		Drive           string `json:"drive"`
		Transmission    string `json:"transmission"`
		Fuel            string `json:"fuel"`
		ManufactureDate struct {
			Year  string `json:"year"`
			Month string `json:"month"`
		} `json:"manufacture_date"`
	} `json:"data"`
	Error string `json:"error"`
}
