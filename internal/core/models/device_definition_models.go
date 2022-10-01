package models

import (
	"encoding/json"
	"math/big"

	"github.com/volatiletech/null/v8"
)

type GetDeviceDefinitionQueryResult struct {
	DeviceDefinitionID string     `json:"deviceDefinitionId"`
	Name               string     `json:"name"`
	ImageURL           string     `json:"imageUrl"`
	Source             string     `json:"source"`
	DeviceMake         DeviceMake `json:"make"`
	Type               DeviceType `json:"type"`
	// VehicleInfo will be empty if not a vehicle type
	VehicleInfo VehicleInfo `json:"vehicleData,omitempty"`
	Metadata    interface{} `json:"metadata"`
	Verified    bool        `json:"verified"`
	// DeviceIntegrations has integrations this vehicle can integrate with, from table device_integrations
	DeviceIntegrations []DeviceIntegration `json:"deviceIntegrations"`
	DeviceStyles       []DeviceStyle       `json:"deviceStyles"`
}

// VehicleInfo represents some standard vehicle specific properties stored in the metadata json field in DB
type VehicleInfo struct {
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
	MakeSlug  string   `json:"makeSlug"`
	ModelSlug string   `json:"modelSlug"`
}

type DeviceIntegration struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Capabilities json.RawMessage `json:"capabilities"`
}

type DeviceStyle struct {
	ID                 string `json:"id"`
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	Name               string `json:"name"`
	ExternalStyleID    string `json:"externalStyleId"`
	Source             string `json:"source"`
	SubModel           string `json:"subModel"`
}

type DeviceMake struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	LogoURL         null.String `json:"logo_url"`
	OemPlatformName null.String `json:"oem_platform_name"`
	TokenID         *big.Int    `json:"tokenId,omitempty"`
	NameSlug        string      `json:"nameSlug"`
}
