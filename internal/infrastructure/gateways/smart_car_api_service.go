package gateways

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	SmartCarVendor = "SmartCar"
)

//go:generate mockgen -source smart_car_api_service.go -destination mocks/smart_car_api_service_mock.go
type SmartCarService interface {
	GetOrCreateSmartCarIntegration(ctx context.Context) (string, error)
	GetSmartCarVehicleData() (*SmartCarCompatibilityData, error)
}

type smartCarService struct {
	baseURL string
	DBS     func() *db.ReaderWriter
	log     zerolog.Logger
}

func NewSmartCarService(dbs func() *db.ReaderWriter, logger zerolog.Logger) SmartCarService {
	return &smartCarService{
		baseURL: "https://api.smartcar.com/v2.0/",
		DBS:     dbs,
		log:     logger,
	}
}

// ParseSmartCarYears parses out the years format in the smartcar document and returns an array of years
func ParseSmartCarYears(yearsPtr *string) ([]int, error) {
	if yearsPtr == nil || len(*yearsPtr) == 0 {
		return nil, errors.New("years string was nil")
	}
	years := *yearsPtr
	if len(years) > 4 {
		var rangeYears []int
		startYear := years[:4]
		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			return nil, errors.Errorf("could not parse start year from: %s", years)
		}
		endYear := time.Now().Year()
		if strings.Contains(years, "-") {
			eyStr := years[5:]
			endYear, err = strconv.Atoi(eyStr)
			if err != nil {
				return nil, errors.Errorf("could not parse end year from: %s", years)
			}
		}
		for y := startYearInt; y <= endYear; y++ {
			rangeYears = append(rangeYears, y)
		}
		return rangeYears, nil
	}
	y, err := strconv.Atoi(years)
	if err != nil {
		return nil, errors.Errorf("could not parse single year from: %s", years)
	}
	return []int{y}, nil
}

func (s *smartCarService) GetOrCreateSmartCarIntegration(ctx context.Context) (string, error) {
	const (
		smartCarType  = "API"
		smartCarStyle = models.IntegrationStyleWebhook
	)
	integration, err := models.Integrations(qm.Where("type = ?", smartCarType),
		qm.And("vendor = ?", SmartCarVendor),
		qm.And("style = ?", smartCarStyle)).One(ctx, s.DBS().Writer)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// create
			integration = &models.Integration{}
			integration.ID = ksuid.New().String()
			integration.Vendor = SmartCarVendor
			integration.Type = smartCarType
			integration.Style = smartCarStyle
			err = integration.Insert(ctx, s.DBS().Writer, boil.Infer())
			if err != nil {
				return "", errors.Wrap(err, "error inserting smart car integration")
			}
		} else {
			return "", errors.Wrap(err, "error fetching smart car integration from database")
		}
	}
	return integration.ID, nil
}

// GetSmartCarVehicleData gets all smartcar data on compatibility from their website
func (s *smartCarService) GetSmartCarVehicleData() (*SmartCarCompatibilityData, error) {
	const url = "https://smartcar.com/page-data/product/compatible-vehicles/page-data.json"
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response from smart car page. status code: %d", res.StatusCode)
	}

	compatibleVehicles := SmartCarCompatibilityData{}
	err = json.NewDecoder(res.Body).Decode(&compatibleVehicles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal json from smart car")
	}
	return &compatibleVehicles, nil
}

type SmartCarCompatibilityData struct {
	ComponentChunkName string `json:"componentChunkName"`
	Path               string `json:"path"`
	Result             struct {
		Data struct {
			AllMakesTable struct {
				Edges []struct {
					Node struct {
						CompatibilityData map[string][]struct {
							Name    string `json:"name"`
							Headers []struct {
								Text    string  `json:"text"`
								Tooltip *string `json:"tooltip"`
							} `json:"headers"`
							Rows [][]struct {
								Color       *string `json:"color"`
								Subtext     *string `json:"subtext"`
								Text        *string `json:"text"`
								Type        *string `json:"type"`
								VehicleType *string `json:"vehicleType"`
							} `json:"rows"`
						} `json:"compatibilityData"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"allMakesTable"`
		} `json:"data"`
	} `json:"result"`
}

// IntegrationCapabilities gets stored on the association table btw a device_definition and the integrations, device_integrations
type IntegrationCapabilities struct {
	Location          bool `json:"location"`
	Odometer          bool `json:"odometer"`
	LockUnlock        bool `json:"lock_unlock"`
	EVBattery         bool `json:"ev_battery"`
	EVChargingStatus  bool `json:"ev_charging_status"`
	EVStartStopCharge bool `json:"ev_start_stop_charge"`
	FuelTank          bool `json:"fuel_tank"`
	TirePressure      bool `json:"tire_pressure"`
	EngineOilLife     bool `json:"engine_oil_life"`
	VehicleAttributes bool `json:"vehicle_attributes"`
	VIN               bool `json:"vin"`
}
