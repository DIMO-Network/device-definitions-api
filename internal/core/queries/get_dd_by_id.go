package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/null/v8"
)

type GetDeviceDefinitionByIDQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetDeviceDefinitionQueryResult struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	Name               string `json:"name"`
	ImageURL           string `json:"imageUrl"`
	// CompatibleIntegrations has systems this vehicle can integrate with
	CompatibleIntegrations []GetDeviceCompatibility `json:"compatibleIntegrations"`
	DeviceMake             DeviceMake               `json:"make"`
	Type                   DeviceType               `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo        GetDeviceVehicleInfo                 `json:"vehicleData,omitempty"`
	Metadata           interface{}                          `json:"metadata"`
	Verified           bool                                 `json:"verified"`
	DeviceIntegrations []GetDeviceDefinitionIntegrationList `json:"deviceIntegrations"`
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

// GetDeviceVehicleInfo represents some standard vehicle specific properties stored in the metadata json field in DB
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
	MPG                 string `json:"mpg,omitempty"`
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

type GetDeviceDefinitionIntegrationList struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

type DeviceMake struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	LogoURL         null.String `json:"logo_url"`
	OemPlatformName null.String `json:"oem_platform_name"`
}

func (*GetDeviceDefinitionByIDQuery) Key() string { return "GetDeviceDefinitionByIdQuery" }

type GetDeviceDefinitionByIDQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionByIDQueryHandler(repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionByIDQueryHandler {
	return GetDeviceDefinitionByIDQueryHandler{
		Repository: repository,
	}
}

func (ch GetDeviceDefinitionByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDQuery)

	dd, _ := ch.Repository.GetByID(ctx, qry.DeviceDefinitionID)

	if dd == nil {
		return nil, &common.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	rp := GetDeviceDefinitionQueryResult{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.String,
		CompatibleIntegrations: []GetDeviceCompatibility{},
		DeviceMake: DeviceMake{
			ID:              dd.R.DeviceMake.ID,
			Name:            dd.R.DeviceMake.Name,
			LogoURL:         dd.R.DeviceMake.LogoURL,
			OemPlatformName: dd.R.DeviceMake.OemPlatformName,
		},
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
		rp.Type.SubModels = common.SubModelsFromStylesDB(dd.R.DeviceStyles)
	}

	// build object for integrations that have all the info
	rp.DeviceIntegrations = []GetDeviceDefinitionIntegrationList{}
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.DeviceIntegrations = append(rp.DeviceIntegrations, GetDeviceDefinitionIntegrationList{
				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendor,
				Region:       di.Region,
				Capabilities: common.JSONOrDefault(di.Capabilities),
			})
		}
	}

	return rp, nil
}
