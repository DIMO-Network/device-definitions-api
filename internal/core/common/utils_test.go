package common

import (
	_ "embed"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
)

func TestBuildExternalIds(t *testing.T) {

	json := null.JSONFrom([]byte(`{"edmunds": "123", "nhtsa": "qwert", "adac": "890" }`))

	got := BuildExternalIds(json)

	assert.Contains(t, got, &coremodels.ExternalID{Vendor: "edmunds", ID: "123"})
	assert.Contains(t, got, &coremodels.ExternalID{Vendor: "nhtsa", ID: "qwert"})
	assert.Contains(t, got, &coremodels.ExternalID{Vendor: "adac", ID: "890"})
}

func TestExternalIdsToGRPC(t *testing.T) {

	extIDs := []*coremodels.ExternalID{
		{Vendor: "edmunds", ID: "123"},
		{Vendor: "nhtsa", ID: "qwert"},
		{Vendor: "adac", ID: "890"},
	}

	got := ExternalIdsToGRPC(extIDs)

	assert.Equal(t, 3, len(got))

	assert.Equal(t, "edmunds", got[0].Vendor)
	assert.Equal(t, "123", got[0].Id)

	assert.Equal(t, "nhtsa", got[1].Vendor)
	assert.Equal(t, "qwert", got[1].Id)

	assert.Equal(t, "adac", got[2].Vendor)
	assert.Equal(t, "890", got[2].Id)
}

//go:embed device_type_vehicle_properties.json
var deviceTypeVehiclePropertyDataSample []byte

func TestBuildDeviceTypeAttributes(t *testing.T) {

	// objective is we feed in db DeviceType of Vehicle with eg our production metadata for attributes
	// we then pass in an array of our update device type attrs with settings for all
	// finally expect that the returned json has all of the update ones, can just use gjson, also it returns as a map but maybe change to just string?
	// edge cases: same value on both properties, repeated properties, properties in attributes by not in the device type attrs.

	// arrange data
	deviceType := &models.DeviceType{
		ID:          ksuid.New().String(),
		Name:        "vehicle",
		Metadatakey: "vehicle_info",
		Properties:  null.JSONFrom(deviceTypeVehiclePropertyDataSample),
	}
	attributes := []*coremodels.UpdateDeviceTypeAttribute{ // these names must match what is in deviceType
		{
			Name:  "fuel_type",
			Value: "gasoline",
		},
		{
			Name:  "driven_wheels",
			Value: "AWD",
		},
		{
			Name:  "number_of_doors",
			Value: "4",
		},
		{
			Name:  "fuel_tank_capacity_gal",
			Value: "22.25",
		},
	}

	got, err := BuildDeviceTypeAttributes(attributes, deviceType)
	require.NoError(t, err)
	// assert
	assert.Equal(t, "gasoline", gjson.GetBytes(got.JSON, "vehicle_info.fuel_type").String())
	assert.Equal(t, "AWD", gjson.GetBytes(got.JSON, "vehicle_info.driven_wheels").String())
	assert.Equal(t, "4", gjson.GetBytes(got.JSON, "vehicle_info.number_of_doors").String())
	assert.Equal(t, "22.25", gjson.GetBytes(got.JSON, "vehicle_info.fuel_tank_capacity_gal").String())
	assert.Equal(t, false, gjson.GetBytes(got.JSON, "vehicle_info.mpg").Exists(), "other properties should not be present")
	//fmt.Printf("got: %s", string(got.JSON))
}

func TestBuildDeviceTypeAttributes_errorsInvalidProperty(t *testing.T) {
	// arrange data
	deviceType := &models.DeviceType{
		ID:          ksuid.New().String(),
		Name:        "vehicle",
		Metadatakey: "vehicle_info",
		Properties:  null.JSONFrom(deviceTypeVehiclePropertyDataSample),
	}
	attributes := []*coremodels.UpdateDeviceTypeAttribute{ // these names must match what is in deviceType
		{
			Name:  "fuel_tank_capacity_gal",
			Value: "22.25",
		},
		{
			Name:  "invalid_property",
			Value: "something",
		},
	}
	// assert
	got, err := BuildDeviceTypeAttributes(attributes, deviceType)
	assert.Equal(t, false, got.Valid)
	assert.ErrorContains(t, err, "invalid", "expected an error when get a property not in device type attrs")
}

func TestBuildDeviceTypeAttributes_noJSONIfNil(t *testing.T) {
	// arrange data
	deviceType := &models.DeviceType{
		ID:          ksuid.New().String(),
		Name:        "vehicle",
		Metadatakey: "vehicle_info",
		Properties:  null.JSONFrom(deviceTypeVehiclePropertyDataSample),
	}
	// assert
	got, err := BuildDeviceTypeAttributes(nil, deviceType)
	require.NoError(t, err)
	assert.Equal(t, false, got.Valid)
	assert.Equal(t, "", string(got.JSON)) // pending to see what this gives
}
