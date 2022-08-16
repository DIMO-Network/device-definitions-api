package queries

import (
	"context"
	"encoding/json"
	"fmt"

	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetByMakeModelYearQuery struct {
	Make  string `json:"make" validate:"required"`
	Model string `json:"model" validate:"required"`
	Year  int    `json:"year" validate:"required"`
}

type GetByMakeModelYearQueryResult struct {
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

func (*GetByMakeModelYearQuery) Key() string { return "GetByMakeModelYearQuery" }

type GetByMakeModelYearQueryHandler struct {
	Repository interfaces.IDeviceDefinitionRepository
}

func NewGetByMakeModelYearQueryHandler(repository interfaces.IDeviceDefinitionRepository) GetByMakeModelYearQueryHandler {
	return GetByMakeModelYearQueryHandler{
		Repository: repository,
	}
}

func (ch GetByMakeModelYearQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetByMakeModelYearQuery)

	dd, _ := ch.Repository.GetByMakeModelAndYears(ctx, qry.Make, qry.Model, qry.Year, true)

	if dd == nil {
		return &GetByMakeModelYearQueryResult{}, nil
	}

	rp := GetByMakeModelYearQueryResult{
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

	return rp, nil
}
