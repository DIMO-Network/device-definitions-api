package gateways

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
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

	urlPath := vincarioPathBuilder(vin, id, va.settings.VincarioAPIKey, va.settings.VincarioAPISecret)
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest(urlPath, "GET", nil)
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

func vincarioPathBuilder(vin, id, key, secret string) string {
	s := vin + "|" + id + "|" + key + "|" + secret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])

	return "/" + key + "/" + controlSum + "/" + id + "/" + vin + ".json"
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
				if fieldValue.Kind() == reflect.Float64 || fieldValue.Kind() == reflect.Float32 {
					f := fieldValue.Float()
					if field.Kind() == reflect.Int {
						field.Set(reflect.ValueOf(int(f)))
					}
				} else {
					return fmt.Errorf("value %v is not assignable to field %s of type %s", value, fieldType.Name, field.Type())
				}
			} else {
				field.Set(fieldValue)
			}
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
		ID    int    `json:"id,omitempty"`
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
	s := strings.TrimSpace(fmt.Sprintf("%s %s %s %d-speed", v.FuelType, v.EngineType, v.Transmission, v.NumberOfGears))
	return s // VehicleID is Make Model Year specific - not style specific.
}

// GetSubModel returns the Body type from Vincario, which we can use as the sub model.
func (v *VincarioInfoResponse) GetSubModel() string {
	return strings.TrimSpace(v.Body)
}
