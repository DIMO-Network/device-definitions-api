package models

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/volatiletech/null/v8"
)

type GetDeviceDefinitionQueryResult struct {
	DeviceDefinitionID string        `json:"deviceDefinitionId"`
	ExternalID         string        `json:"external_id"`
	Name               string        `json:"name"`
	ImageURL           string        `json:"imageUrl"`
	Source             string        `json:"source"`
	HardwareTemplateID string        `json:"hardware_template_id"`
	DeviceMake         DeviceMake    `json:"make"`
	Type               DeviceType    `json:"type"`
	VehicleInfo        VehicleInfo   `json:"vehicleData,omitempty"`
	Metadata           interface{}   `json:"metadata"`
	Verified           bool          `json:"verified"`
	ExternalIds        []*ExternalID `json:"externalIds"`
	// DeviceIntegrations has integrations this vehicle can integrate with, from table device_integrations
	DeviceIntegrations     []DeviceIntegration   `json:"deviceIntegrations"`
	CompatibleIntegrations []DeviceIntegration   `json:"compatibleIntegrations"`
	DeviceStyles           []DeviceStyle         `json:"deviceStyles"`
	DeviceAttributes       []DeviceTypeAttribute `json:"deviceAttributes"`
}

type DeviceTypeAttribute struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Value       string   `json:"value"`
	Option      []string `json:"options"`
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
	ID       string                     `json:"id"`
	Type     string                     `json:"type"`
	Style    string                     `json:"style"`
	Vendor   string                     `json:"vendor"`
	Region   string                     `json:"region"`
	Features []DeviceIntegrationFeature `json:"features"`
}

type DeviceIntegrationFeature struct {
	FeatureKey   string `json:"featureKey"`
	SupportLevel int    `json:"supportLevel"`
}

type DeviceStyle struct {
	ID                 string                `json:"id"`
	DeviceDefinitionID string                `json:"deviceDefinitionId"`
	Name               string                `json:"name"`
	ExternalStyleID    string                `json:"externalStyleId"`
	Source             string                `json:"source"`
	SubModel           string                `json:"subModel"`
	HardwareTemplateID string                `json:"hardware_template_id"`
	Metadata           []DeviceTypeAttribute `json:"metadata"`
}

type DeviceMake struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	LogoURL            null.String         `json:"logo_url"`
	OemPlatformName    null.String         `json:"oem_platform_name"`
	TokenID            *big.Int            `json:"tokenId,omitempty"`
	NameSlug           string              `json:"nameSlug"`
	ExternalIds        json.RawMessage     `json:"external_ids"`
	ExternalIdsTyped   []*ExternalID       `json:"externalIdsTyped"`
	Metadata           json.RawMessage     `json:"metadata"`
	MetadataTyped      *DeviceMakeMetadata `json:"metadataTyped"`
	HardwareTemplateID null.String         `json:"hardware_template_id"`
	CreatedAt          time.Time           `json:"created_at,omitempty"`
	UpdatedAt          time.Time           `json:"updated_at,omitempty"`
}

type ExternalID struct {
	Vendor string `json:"vendor"`
	ID     string `json:"id"`
}

type DeviceMakeMetadata struct {
	RideGuideLink string `json:"ride_guide_link"`
}

type GetDeviceDefinitionHardwareTemplateQueryResult struct {
	TemplateID string `json:"template_id"`
}
