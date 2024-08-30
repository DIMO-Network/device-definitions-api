package gateways

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
)

//go:generate mockgen -source drivly_api_service.go -destination mocks/drivly_api_service_mock.go -package mocks
type DrivlyAPIService interface {
	GetVINInfo(vin string) (*DrivlyVINResponse, error)
}

type drivlyAPIService struct {
	Settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
}

func NewDrivlyAPIService(settings *config.Settings) DrivlyAPIService {
	if settings.DrivlyAPIKey == "" {
		panic("Drivly configuration not set")
	}
	h := map[string]string{"x-api-key": settings.DrivlyAPIKey}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.DrivlyVINAPIURL.String(), "", 10*time.Second, h, true, shared.WithRetry(1))

	return &drivlyAPIService{
		Settings:      settings,
		httpClientVIN: hcwv,
	}
}

// GetVINInfo is the basic enriched VIN call, that is pretty standard now. Looks in multiple sources in their backend.
func (ds *drivlyAPIService) GetVINInfo(vin string) (*DrivlyVINResponse, error) {
	res, err := executeAPI(ds.httpClientVIN, fmt.Sprintf("/api/%s/", vin))
	if err != nil {
		return nil, err
	}
	v := &DrivlyVINResponse{}
	err = json.Unmarshal(res, v)
	if err != nil {
		return nil, err
	}

	if v.Year == "0" || len(v.Year) == 0 || len(v.Model) == 0 || len(v.Make) == 0 {
		return nil, fmt.Errorf("decode failed due to invalid MMY")
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
		return nil, errors.Wrapf(err, "error calling api => %s", path)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	return body, nil
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
