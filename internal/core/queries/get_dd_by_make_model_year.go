package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByMakeModelYearQuery struct {
	Make  string `json:"make" validate:"required"`
	Model string `json:"model" validate:"required"`
	Year  int    `json:"year" validate:"required"`
}

type GetDeviceDefinitionByMakeModelYearQueryResult struct {
	DeviceDefinitionID string  `json:"deviceDefinitionId"`
	Name               string  `json:"name"`
	ImageURL           *string `json:"imageUrl"`
	// CompatibleIntegrations has systems this vehicle can integrate with
	CompatibleIntegrations []GetDeviceCompatibility `json:"compatibleIntegrations"`
	Type                   DeviceType               `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo GetDeviceVehicleInfo `json:"vehicleData,omitempty"`
	Metadata    interface{}          `json:"metadata"`
	Verified    bool                 `json:"verified"`
}

// DeviceCompatibility represents what systems we know this is compatible with
type GetDeviceCompatibility struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

// DeviceVehicleInfo represents some standard vehicle specific properties stored in the metadata json field in DB
type GetDeviceVehicleInfo struct {
	FuelType            string `json:"fuel_type,omitempty"`
	DrivenWheels        string `json:"driven_wheels,omitempty"`
	NumberOfDoors       string `json:"number_of_doors,omitempty"`
	BaseMSRP            int    `json:"base_msrp,omitempty"`
	EPAClass            string `json:"epa_class,omitempty"`
	VehicleType         string `json:"vehicle_type,omitempty"` // VehicleType PASSENGER CAR, from NHTSA
	MPGHighway          string `json:"mpg_highway,omitempty"`
	MPGCity             string `json:"mpg_city,omitempty"`
	FuelTankCapacityGal string `json:"fuel_tank_capacity_gal,omitempty"`
}

// DeviceType whether it is a vehicle or other type and basic information
type DeviceType struct {
	// Type is eg. Vehicle, E-bike, roomba
	Type      string   `json:"type"`
	Make      string   `json:"make"`
	Model     string   `json:"model"`
	Year      int      `json:"year"`
	SubModels []string `json:"subModels"`
}

func (*GetDeviceDefinitionByMakeModelYearQuery) Key() string {
	return "GetDeviceDefinitionByMakeModelYearQuery"
}

type GetDeviceDefinitionByMakeModelYearQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionByMakeModelYearQueryHandler(repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionByMakeModelYearQueryHandler {
	return GetDeviceDefinitionByMakeModelYearQueryHandler{
		Repository: repository,
	}
}

func (ch GetDeviceDefinitionByMakeModelYearQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByMakeModelYearQuery)

	dd, _ := ch.Repository.GetByMakeModelAndYears(ctx, qry.Make, qry.Model, qry.Year, true)

	if dd == nil {
		return &GetDeviceDefinitionByMakeModelYearQueryResult{}, nil
	}

	rp := GetDeviceDefinitionByMakeModelYearQueryResult{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.Ptr(),
		CompatibleIntegrations: []GetDeviceCompatibility{},
		Type: DeviceType{
			Type:  "Vehicle",
			Make:  dd.R.DeviceMake.Name,
			Model: dd.Model,
			Year:  int(dd.Year),
		},
		Metadata: string(dd.Metadata.JSON),
		Verified: dd.Verified,
	}

	// vehicle info
	var vi map[string]GetDeviceVehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		rp.VehicleInfo = vi["vehicle_info"]
	}

	if dd.R != nil {
		// compatible integrations
		rp.CompatibleIntegrations = deviceCompatibilityFromDB(dd.R.DeviceIntegrations)
		// sub_models
		rp.Type.SubModels = subModelsFromStylesDB(dd.R.DeviceStyles)
	}

	return rp, nil
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func deviceCompatibilityFromDB(dbDIS models.DeviceIntegrationSlice) []GetDeviceCompatibility {
	if len(dbDIS) == 0 {
		return []GetDeviceCompatibility{}
	}
	compatibilities := make([]GetDeviceCompatibility, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = GetDeviceCompatibility{
			ID:           di.IntegrationID,
			Type:         di.R.Integration.Type,
			Style:        di.R.Integration.Style,
			Vendor:       di.R.Integration.Vendor,
			Region:       di.Region,
			Capabilities: common.JSONOrDefault(di.Capabilities),
		}
	}
	return compatibilities
}

// SubModelsFromStylesDB gets the unique style.SubModel from the styles slice, deduping sub_model
func subModelsFromStylesDB(styles models.DeviceStyleSlice) []string {
	items := map[string]string{}
	for _, style := range styles {
		if _, ok := items[style.SubModel]; !ok {
			items[style.SubModel] = style.Name
		}
	}

	sm := make([]string, len(items))
	i := 0
	for key := range items {
		sm[i] = key
		i++
	}
	sort.Strings(sm)
	return sm
}
