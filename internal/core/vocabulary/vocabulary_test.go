package vocabulary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(f float64) *float64 { return &f }

// A trimmed stand-in for /schema/device-type-vehicle.json, carrying one
// attribute of each type plus the four the mapping tables actually act on.
func testVocab() *DeviceType {
	return &DeviceType{
		ID: "vehicle",
		Attributes: []Attribute{
			{Name: "powertrain_type", Type: "enum", Options: []string{"ICE", "HEV", "PHEV", "BEV", "FCEV"}},
			{Name: "fuel_type", Type: "enum", Options: []string{"gasoline", "diesel", "electric", "flex_fuel", "hydrogen", "cng", "lpg"}},
			{Name: "driven_wheels", Type: "enum", Options: []string{"FWD", "RWD", "AWD", "4WD"}},
			{Name: "vehicle_type", Type: "enum", Options: []string{"sedan", "hatchback", "wagon", "coupe", "convertible", "suv", "minivan", "van", "pickup", "truck", "motorcycle", "chassis_cab"}},
			{Name: "emissions_standard", Type: "enum", Options: []string{"euro_4", "euro_5", "euro_6", "euro_6d"}},
			{Name: "number_of_doors", Type: "integer", Minimum: ptr(1), Maximum: ptr(8)},
			{Name: "fuel_tank_capacity_gal", Type: "number", Minimum: ptr(0), Maximum: ptr(100)},
			{Name: "wheelbase_in", Type: "number", Minimum: ptr(0), Maximum: ptr(500)},
			{Name: "manufacturer_code", Type: "string"},
		},
	}
}

func TestMapAttributesTypesValues(t *testing.T) {
	// The contract requires typed values: fuel_tank_capacity_gal is 15.8, not
	// "15.800000". 6% of the old catalog's rows were "" or "<nil>".
	mapped, dropped := MapAttributes(map[string]string{
		"fuel_tank_capacity_gal": "15.800000",
		"number_of_doors":        "4",
		"manufacturer_code":      "2532",
	}, testVocab())

	assert.Empty(t, dropped)
	assert.Equal(t, 15.8, mapped["fuel_tank_capacity_gal"])
	assert.Equal(t, 4, mapped["number_of_doors"])
	assert.Equal(t, "2532", mapped["manufacturer_code"])
}

func TestMapAttributesOmitsDeadPlaceholders(t *testing.T) {
	mapped, dropped := MapAttributes(map[string]string{
		"fuel_type":       "",
		"driven_wheels":   "<nil>",
		"number_of_doors": "N/A",
		"vehicle_type":    "null",
	}, testVocab())

	// Absent, never empty: these are unrepresentable in the contract, and they
	// are not drops either -- there was nothing there to carry.
	assert.Empty(t, mapped)
	assert.Empty(t, dropped)
}

func TestMapAttributesFoldsSpellings(t *testing.T) {
	mapped, dropped := MapAttributes(map[string]string{
		"fuel_type":     "Petrol",
		"driven_wheels": "All-Wheel Drive",
		"vehicle_type":  "Crossover Utility Vehicle (CUV)",
		"epa_class":     "Euro 6d-temp",
	}, testVocab())

	require.Empty(t, dropped)
	assert.Equal(t, "gasoline", mapped["fuel_type"])
	assert.Equal(t, "AWD", mapped["driven_wheels"])
	assert.Equal(t, "suv", mapped["vehicle_type"])
	// epa_class is renamed: the unit and the axis belong to the vocabulary.
	assert.Equal(t, "euro_6d", mapped["emissions_standard"])
	assert.NotContains(t, mapped, "epa_class")
}

// "Hybrid" arrives as a raw fuel_type. Powertrain derivation has already made
// it HEV; a hybrid burns gasoline, so folding the leftover string onto its
// actual fuel is what stops the record claiming a fuel the vocabulary lacks.
// This conflation is what produced the blended Camry.
func TestMapAttributesFoldsHybridFuelOntoGasoline(t *testing.T) {
	mapped, dropped := MapAttributes(map[string]string{"fuel_type": "Hybrid"}, testVocab())
	assert.Empty(t, dropped)
	assert.Equal(t, "gasoline", mapped["fuel_type"])
}

// A drop must say which decision it was. Routing these through the generic
// reason would make a deliberate choice look like a typo to everyone reading
// the report rather than the source.
func TestMapAttributesReportsDeliberateDropsDistinctly(t *testing.T) {
	_, dropped := MapAttributes(map[string]string{
		"driven_wheels": "4x2",
		"epa_class":     "6L",
	}, testVocab())

	require.Len(t, dropped, 2)
	byName := map[string]DroppedAttribute{}
	for _, d := range dropped {
		byName[d.Name] = d
	}
	assert.Contains(t, byName["driven_wheels"].Reason, "driven wheel count")
	assert.Contains(t, byName["epa_class"].Reason, "coding scheme")
	// Distinct from the generic reason, which is the whole point.
	assert.NotContains(t, byName["driven_wheels"].Reason, "no mapping")
	assert.NotContains(t, byName["epa_class"].Reason, "no mapping")
}

func TestMapAttributesReportsWhatItDidNotCarry(t *testing.T) {
	mapped, dropped := MapAttributes(map[string]string{
		"generation":             "6",
		"epa_class":              "Calss III",
		"number_of_doors":        "4.5",
		"fuel_tank_capacity_gal": "unknown",
	}, testVocab())

	assert.Empty(t, mapped)
	require.Len(t, dropped, 4)
	reasons := map[string]string{}
	for _, d := range dropped {
		reasons[d.Name] = d.Reason
	}
	assert.Equal(t, "no such attribute in the vocabulary", reasons["generation"])
	assert.Equal(t, "value not in the vocabulary and no mapping", reasons["epa_class"])
	assert.Equal(t, "not an integer", reasons["number_of_doors"])
	assert.Equal(t, "not a number", reasons["fuel_tank_capacity_gal"])
}

func TestMapAttributesEnforcesVocabularyRange(t *testing.T) {
	_, dropped := MapAttributes(map[string]string{"number_of_doors": "12"}, testVocab())
	require.Len(t, dropped, 1)
	assert.Equal(t, "above the vocabulary range", dropped[0].Reason)
}

func TestMapAttributesParsesWheelbase(t *testing.T) {
	mapped, dropped := MapAttributes(map[string]string{"wheelbase": "111.2 in"}, testVocab())
	assert.Empty(t, dropped)
	assert.Equal(t, 111.2, mapped["wheelbase_in"])
}

// The drop report is read by a human deciding what to fix. Map iteration order
// would otherwise reshuffle it on every call and make two runs look different.
func TestMapAttributesDropReportIsStable(t *testing.T) {
	in := map[string]string{"zulu": "1", "alpha": "2", "mike": "3", "bravo": "4"}
	first, _ := MapAttributes(in, testVocab())
	assert.Empty(t, first)
	_, dropped := MapAttributes(in, testVocab())
	names := make([]string, len(dropped))
	for i, d := range dropped {
		names[i] = d.Name
	}
	assert.Equal(t, []string{"alpha", "bravo", "mike", "zulu"}, names)
}

func TestMapAttributesWithoutVocabularyCarriesNothing(t *testing.T) {
	mapped, dropped := MapAttributes(map[string]string{"fuel_type": "Petrol"}, nil)
	assert.Empty(t, mapped)
	assert.Empty(t, dropped)
}
