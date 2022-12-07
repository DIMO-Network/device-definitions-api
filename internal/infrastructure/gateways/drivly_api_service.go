package gateways

import (
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
	"io"
	"time"
)

//go:generate mockgen -source drivly_api_service.go -destination mocks/drivly_api_service_mock.go
type DrivlyAPIService interface {
	GetVINInfo(vin string) (*VINInfoResponse, error)
	//GetEdmundsByVIN(vin string) (map[string]interface{}, error)
}

type drivlyAPIService struct {
	Settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
}

func NewDrivlyAPIService(settings *config.Settings) DrivlyAPIService {
	if settings.DrivlyVINAPIURL == "" || settings.DrivlyAPIKey == "" {
		panic("Drivly configuration not set")
	}
	h := map[string]string{"x-api-key": settings.DrivlyAPIKey}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.DrivlyVINAPIURL, "", 10*time.Second, h, true)

	return &drivlyAPIService{
		Settings:      settings,
		httpClientVIN: hcwv,
	}
}

// GetVINInfo is the basic enriched VIN call, that is pretty standard now. Looks in multiple sources in their backend.
func (ds *drivlyAPIService) GetVINInfo(vin string) (*VINInfoResponse, error) {
	res, err := executeAPI(ds.httpClientVIN, fmt.Sprintf("/api/%s/", vin))
	if err != nil {
		return nil, err
	}
	v := &VINInfoResponse{}
	err = json.Unmarshal(res, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func executeAPI(httpClient shared.HTTPClientWrapper, path string) ([]byte, error) {
	res, err := httpClient.ExecuteRequest(path, "GET", nil)
	if res == nil {
		if err != nil {
			return nil, errors.Wrapf(err, "error calling driv.ly api => %s", path)
		}
		return nil, fmt.Errorf("received error with no response when calling GET to %s", path)
	}

	if err != nil && res.StatusCode != 404 {
		return nil, errors.Wrapf(err, "error calling driv.ly api => %s", path)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	return body, nil
}

type VINInfoResponse struct {
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
	MsrpBase                 int      `json:"msrpBase"`
	MsrpDiscount             int      `json:"msrpDiscount"`
	MsrpOptions              int      `json:"msrpOptions"`
	MsrpDelivery             int      `json:"msrpDelivery"`
	Msrp                     int      `json:"msrp"`
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
