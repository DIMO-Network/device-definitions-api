//nolint:tagliatelle
package models

import (
	"encoding/json"
)

type GetDeviceDefinitionQueryResult struct {
	DeviceDefinitionID string                      `json:"deviceDefinitionId"`
	NameSlug           string                      `json:"nameSlug"`
	Name               string                      `json:"name"`
	ImageURL           string                      `json:"imageUrl"`
	HardwareTemplateID string                      `json:"hardware_template_id"`
	Metadata           []byte                      `json:"metadata"`
	Verified           bool                        `json:"verified"`
	DeviceStyles       []DeviceStyle               `json:"deviceStyles"`
	DeviceAttributes   []DeviceTypeAttributeEditor `json:"deviceAttributes"`
	Transactions       []string                    `json:"transactions"`
	MakeName           string                      `json:"makeSlug"`
	MakeTokenID        int                         `json:"makeTokenId"`
}

type DeviceTypeAttributeEditor struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Value       string   `json:"value"`
	Option      []string `json:"options"`
}

// nolint:tagliatelle

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
	ID                 string                      `json:"id"`
	DefinitionID       string                      `json:"definitionId"`
	Name               string                      `json:"name"`
	ExternalStyleID    string                      `json:"externalStyleId"`
	Source             string                      `json:"source"`
	SubModel           string                      `json:"subModel"`
	HardwareTemplateID string                      `json:"hardware_template_id"`
	Metadata           []DeviceTypeAttributeEditor `json:"metadata"`
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

// DeviceDefinitionTablelandModel model returned by on-chain sql lite from tableland
type DeviceDefinitionTablelandModel struct {
	ID         string                    `json:"id"`
	KSUID      string                    `json:"ksuid"`
	Model      string                    `json:"model"`
	Year       int                       `json:"year"`
	DeviceType string                    `json:"devicetype"`
	ImageURI   string                    `json:"imageuri"`
	Metadata   *DeviceDefinitionMetadata `json:"metadata"`
}

// DeviceDefinitionMetadata part of tableland DD: includes a list of device-specific attributes.
type DeviceDefinitionMetadata struct {
	DeviceAttributes []DeviceTypeAttribute `json:"device_attributes"`
}

// DeviceTypeAttribute part of tableland DD: name and value key pair of attributes
type DeviceTypeAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// UnmarshalJSON customizes the unmarshaling of DeviceDefinitionTablelandModel to handle cases where metadata is an empty string.
func (d *DeviceDefinitionTablelandModel) UnmarshalJSON(data []byte) error {
	type Alias DeviceDefinitionTablelandModel // Create an alias to avoid recursion

	aux := &struct {
		Metadata json.RawMessage `json:"metadata"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Metadata) > 0 && string(aux.Metadata) != `""` {
		metadata := new(DeviceDefinitionMetadata)
		if err := json.Unmarshal(aux.Metadata, metadata); err != nil {
			return err
		}
		d.Metadata = metadata
	}

	return nil
}
