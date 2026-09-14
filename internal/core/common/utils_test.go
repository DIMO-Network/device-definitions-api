package common

import (
	_ "embed"
	"regexp"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	"github.com/aarondl/null/v8"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestBuildExternalIds(t *testing.T) {

	json := null.JSONFrom([]byte(`{"edmunds": "123", "nhtsa": "qwert", "adac": "890" }`))

	got := BuildExternalIDs(json)

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

	got := ExternalIDsToGRPC(extIDs)

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
	require.NotNil(t, got)
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

func TestDeviceDefinitionSlug(t *testing.T) {
	tests := []struct {
		makeSlug  string
		modelSlug string
		year      int16
		want      string
	}{
		{
			makeSlug:  "audi",
			modelSlug: "tt,-tts",
			year:      2010,
			want:      "audi_tt-tts_2010",
		},
		{
			makeSlug:  "mercedes-benz",
			modelSlug: "v,-v-class",
			year:      2023,
			want:      "mercedes-benz_v-v-class_2023",
		},
		{
			makeSlug:  "mercedes-benz",
			modelSlug: "v-class,-vito,-vito-tourer",
			year:      2023,
			want:      "mercedes-benz_v-class-vito-vito-tourer_2023",
		},
		{
			makeSlug:  "chrysler",
			modelSlug: "pacifica/voyager",
			year:      2018,
			want:      "chrysler_pacifica-voyager_2018",
		},
		{
			makeSlug:  "volkswagen",
			modelSlug: "id.4,-id.5",
			year:      2023,
			want:      "volkswagen_id-4-id-5_2023",
		},
	}
	for _, tt := range tests {
		t.Run(tt.makeSlug+""+tt.modelSlug, func(t *testing.T) {
			assert.Equalf(t, tt.want, DeviceDefinitionSlug(tt.makeSlug, tt.modelSlug, tt.year), "SlugString(%v)", tt.makeSlug)
		})
	}
}

// workerIDRegex mirrors definitions-worker/src/template.ts ID_RE. The worker
// answers 422 for any id outside it, and dd-api has no other way to learn
// that: a decode that builds such an id fails on every retry, forever.
var workerIDRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9._&+-]*_[a-z0-9._&+-]+_[0-9]{4}$`)

func TestDeviceDefinitionSlugMatchesWorkerIDRegex(t *testing.T) {
	tests := []struct {
		make  string
		model string
		year  int16
		want  string
	}{
		// The two shapes seen in production decodes.
		{make: "Volkswagen", model: "up!", year: 2025, want: "volkswagen_up_2025"},
		{make: "Subaru", model: "Tribeca (NY/NJ)", year: 2008, want: "subaru_tribeca-ny-nj_2008"},
		{make: "Ford", model: `"Special" Edition`, year: 2019, want: "ford_special-edition_2019"},
		{make: "Tesla", model: "Model 3", year: 2022, want: "tesla_model-3_2022"},
		{make: "Chrysler", model: "C/D 4.5", year: 2001, want: "chrysler_c-d-4-5_2001"},
		// Characters the worker accepts survive.
		{make: "Dodge", model: "Town & Country", year: 2012, want: "dodge_town-&-country_2012"},
		{make: "Mercedes-Benz", model: "GLE 450+", year: 2024, want: "mercedes-benz_gle-450+_2024"},
		// Stripping must not leave a run of dashes or a dangling one.
		{make: "Kia", model: "Soul !EV!", year: 2020, want: "kia_soul-ev_2020"},
	}
	for _, tt := range tests {
		t.Run(tt.make+" "+tt.model, func(t *testing.T) {
			// The decode path slugs both parts first; the sanitizer must
			// hold for what SlugString leaves behind.
			got := DeviceDefinitionSlug(stringutils.SlugString(tt.make), stringutils.SlugString(tt.model), tt.year)
			assert.Equal(t, tt.want, got)
			assert.Regexp(t, workerIDRegex, got)

			// cmd/device-definitions-api/decode_vin.go passes the decoded
			// make and model raw, so the raw shape must be safe as well.
			raw := DeviceDefinitionSlug(tt.make, tt.model, tt.year)
			assert.Equal(t, tt.want, raw)
			assert.Regexp(t, workerIDRegex, raw)
		})
	}
}
