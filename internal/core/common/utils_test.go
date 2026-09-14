package common

import (
	_ "embed"
	"regexp"
	"testing"

	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

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
		// A stripped character leaves its neighbours as they were.
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

// legacyDeviceDefinitionSlug is DeviceDefinitionSlug as it was before ids were
// checked against the worker. Every id it builds that the worker accepts must
// come out of DeviceDefinitionSlug unchanged, or new decodes of that model stop
// finding their template and create a duplicate.
func legacyDeviceDefinitionSlug(makeSlug, modelSlug string, year int16) string {
	modelSlugCleaned := strings.ReplaceAll(modelSlug, ",", "")
	modelSlugCleaned = strings.ReplaceAll(modelSlugCleaned, "/", "-")
	modelSlugCleaned = strings.ReplaceAll(modelSlugCleaned, ".", "-")
	return fmt.Sprintf("%s_%s_%d", makeSlug, modelSlugCleaned, year)
}

func TestDeviceDefinitionSlugKeepsIDsTheWorkerAccepts(t *testing.T) {
	// Live catalog ids whose model slug carries a dash run or a trailing dash.
	tests := []struct {
		make  string
		model string
		year  int16
	}{
		{"volkswagen", "id--buzz", 2024},
		{"volkswagen", "id--buzz-cargo", 2025},
		{"ford", "ranger---ra", 2022},
		{"bmw", "x3-", 2026},
	}
	for _, tt := range tests {
		want := fmt.Sprintf("%s_%s_%d", tt.make, tt.model, tt.year)
		t.Run(want, func(t *testing.T) {
			assert.Equal(t, want, DeviceDefinitionSlug(tt.make, tt.model, tt.year))
		})
	}
	// The decode path slugs "ID. Buzz" to id--buzz; the raw cmd path must agree.
	assert.Equal(t, "volkswagen_id--buzz_2024", DeviceDefinitionSlug(stringutils.SlugString("Volkswagen"), stringutils.SlugString("ID. Buzz"), 2024))
	assert.Equal(t, "volkswagen_id--buzz_2024", DeviceDefinitionSlug("Volkswagen", "ID. Buzz", 2024))
}

// TestDeviceDefinitionSlugAgainstManifest replays every id in a definitions
// manifest through DeviceDefinitionSlug. Run it with
// DEFINITIONS_MANIFEST=path/to/manifest.json (for example a download of
// https://definitions.dimo.org/manifest.json). An id the legacy builder
// produces and the worker accepts must be unchanged; one the worker refuses
// must be repaired into one it accepts.
func TestDeviceDefinitionSlugAgainstManifest(t *testing.T) {
	path := os.Getenv("DEFINITIONS_MANIFEST")
	if path == "" {
		t.Skip("DEFINITIONS_MANIFEST not set")
	}
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	ids := manifestIDs(t, raw)
	require.NotEmpty(t, ids)

	var checked, unchanged, repaired int
	for _, id := range ids {
		first, last := strings.Index(id, "_"), strings.LastIndex(id, "_")
		if first <= 0 || last <= first {
			continue
		}
		year, err := strconv.Atoi(id[last+1:])
		if err != nil {
			continue
		}
		mk, model := id[:first], id[first+1:last]
		legacy := legacyDeviceDefinitionSlug(mk, model, int16(year))
		got := DeviceDefinitionSlug(mk, model, int16(year))
		checked++
		if workerIDRegex.MatchString(legacy) {
			if assert.Equal(t, legacy, got, "an id the worker accepts changed: %s", id) {
				unchanged++
			}
			continue
		}
		if assert.Regexp(t, workerIDRegex, got, "a refused id was not repaired: %s", id) {
			repaired++
		}
	}
	t.Logf("manifest ids=%d checked=%d unchanged=%d repaired=%d", len(ids), checked, unchanged, repaired)
}

// manifestIDs reads the definition ids from a manifest whose definitions are
// either a list of objects with an id or an object keyed by id.
func manifestIDs(t *testing.T, raw []byte) []string {
	var doc struct {
		Definitions json.RawMessage `json:"definitions"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))
	var list []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(doc.Definitions, &list); err == nil {
		ids := make([]string, 0, len(list))
		for _, d := range list {
			if d.ID != "" {
				ids = append(ids, d.ID)
			}
		}
		return ids
	}
	var byID map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(doc.Definitions, &byID))
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	return ids
}

// manifestDefinition is one catalog definition as the manifest carries it: the
// stored id, and the manufacturer name, model and year a decode builds it from.
type manifestDefinition struct {
	ID           string `json:"id"`
	Manufacturer struct {
		Name string `json:"name"`
	} `json:"manufacturer"`
	Model string `json:"model"`
	Year  int    `json:"year"`
}

func manifestDefinitions(t *testing.T, raw []byte) []manifestDefinition {
	var doc struct {
		Definitions []manifestDefinition `json:"definitions"`
	}
	require.NoError(t, json.Unmarshal(raw, &doc))
	return doc.Definitions
}

// TestDeviceDefinitionSlugAgainstManifestNames builds every definition's id
// from its real manufacturer name and model, the inputs a decode starts from,
// through both call shapes: slugged first, as the decode path passes them, and
// raw, as cmd/device-definitions-api does. The shapes must agree, an id the
// legacy builder produced that the worker accepts must be unchanged, and a
// refused one must be repaired into one the worker accepts. Feeding a stored
// id's own parts back in cannot show this: the legacy builder returns most of
// those unchanged by construction. Run with DEFINITIONS_MANIFEST as above.
func TestDeviceDefinitionSlugAgainstManifestNames(t *testing.T) {
	path := os.Getenv("DEFINITIONS_MANIFEST")
	if path == "" {
		t.Skip("DEFINITIONS_MANIFEST not set")
	}
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	defs := manifestDefinitions(t, raw)
	require.NotEmpty(t, defs)

	var checked, agree, unchanged, repaired, newRebuildsStored, legacyRebuildsStored int
	for _, d := range defs {
		name := d.Manufacturer.Name
		if name == "" || d.Model == "" || d.Year <= 0 || d.Year > math.MaxInt16 {
			continue
		}
		year := int16(d.Year)
		mk, model := stringutils.SlugString(name), stringutils.SlugString(d.Model)
		legacy := legacyDeviceDefinitionSlug(mk, model, year)
		slugged := DeviceDefinitionSlug(mk, model, year)
		rawID := DeviceDefinitionSlug(name, d.Model, year)
		checked++
		if assert.Equal(t, slugged, rawID, "raw and slugged inputs disagree for %s (%q %q)", d.ID, name, d.Model) {
			agree++
		}
		if workerIDRegex.MatchString(legacy) {
			if assert.Equal(t, legacy, slugged, "an id the worker accepts changed for %s (%q %q)", d.ID, name, d.Model) {
				unchanged++
			}
		} else if assert.Regexp(t, workerIDRegex, slugged, "a refused id was not repaired for %s (%q %q)", d.ID, name, d.Model) {
			repaired++
		}
		if slugged == d.ID {
			newRebuildsStored++
		}
		if legacy == d.ID {
			legacyRebuildsStored++
		}
	}
	t.Logf("definitions=%d checked=%d raw-agrees=%d valid-unchanged=%d repaired=%d rebuilds-stored-id new=%d legacy=%d",
		len(defs), checked, agree, unchanged, repaired, newRebuildsStored, legacyRebuildsStored)
}
