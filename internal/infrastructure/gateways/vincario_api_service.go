package gateways

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
	"reflect"
	"time"
)

//go:generate mockgen -source vincario_api_service.go -destination mocks/vincario_api_service_mock.go -package mocks
type VincarioAPIService interface {
	DecodeVIN(vin string) (*VincarioInfoResponse, error)
}

type vincarioAPIService struct {
	settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
	log           *zerolog.Logger
}

func NewVincarioAPIService(settings *config.Settings, log *zerolog.Logger) VincarioAPIService {
	if settings.VincarioAPIURL == "" || settings.VincarioAPISecret == "" {
		panic("Vincario configuration not set")
	}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.VincarioAPIURL, "", 10*time.Second, nil, false)

	return &vincarioAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
		log:           log,
	}
}

func (va *vincarioAPIService) DecodeVIN(vin string) (*VincarioInfoResponse, error) {
	id := "decode"

	s := vin + "|" + id + "|" + va.settings.VincarioAPIKey + "|" + va.settings.VincarioAPISecret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest("/"+va.settings.VincarioAPIKey+"/"+controlSum+"/"+id+"/"+vin+".json", "GET", nil)
	if err != nil {
		return nil, err
	}

	// decode JSON from response body
	var data tempResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	infoResp := VincarioInfoResponse{}
	for _, s2 := range data.Decode {
		// find the property in the struct with the label name in the key metadata
		err := setStructPropertiesByMetadataKey(&infoResp, s2.Label, s2.Value)
		if err != nil {
			va.log.Warn().Err(err)
		}
	}

	return &infoResp, nil
}

func setStructPropertiesByMetadataKey(structPtr interface{}, key string, value interface{}) error {
	structValue := reflect.ValueOf(structPtr).Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		if fieldType.Tag.Get("key") == key {
			if !field.CanSet() {
				return fmt.Errorf("field %s is unexported and cannot be set", fieldType.Name)
			}

			fieldValue := reflect.ValueOf(value)

			if !fieldValue.Type().AssignableTo(field.Type()) {
				return fmt.Errorf("value %v is not assignable to field %s of type %s", value, fieldType.Name, field.Type())
			}

			field.Set(fieldValue)
			return nil
		}
	}

	return fmt.Errorf("struct does not contain a field with key %s", key)
}

type tempResponse struct {
	Balance struct {
		APIDecode int `json:"API Decode"`
	} `json:"balance"`
	Decode []struct {
		Label string `json:"label"`
		Value any    `json:"value"`
		Id    int    `json:"id,omitempty"`
	} `json:"decode"`
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
	AverageCO2Emission       float64  `key:"Average CO2 Emission (g\/km)"`
	NumberOfWheels           int      `key:"Number Wheels"`
	NumberOfAxles            int      `key:"Number of Axles"`
	NumberOfDoors            int      `key:"Number of Doors"`
	PowerSteering            string   `key:"Power Steering"`
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
	WidthIncludingMirrors    int      `key:"Width including mirrors (mm)"`
	RearOverhang             int      `key:"Rear Overhang (mm)"`
	FrontOverhang            int      `key:"Front Overhang (mm)"`
	TrackFront               int      `key:"Track Front (mm)"`
	TrackRear                int      `key:"Track Rear (mm)"`
	MaxSpeed                 int      `key:"Max Speed (km\/h)"`
	WeightEmpty              int      `key:"Weight Empty (kg)"`
	MaxWeight                int      `key:"Max Weight (kg)"`
	MaxRoofLoad              int      `key:"Max roof load (kg)"`
	TrailerLoadWithoutBrakes int      `key:"Permitted trailer load without brakes (kg)"`
	ABS                      int      `key:"ABS"`
	CheckDigit               string   `key:"Check Digit"`
	SequentialNumber         string   `key:"Sequential Number"`
}
