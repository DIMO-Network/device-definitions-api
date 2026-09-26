package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Template attributes decode into map[string]any, so every number is a
// float64. fmt.Sprint formats those with %g -- a price of 1250000 renders as
// "1.25e+06" and 15.8 as "1.58e+01" -- and ranging over the map puts the
// attributes in a different order on every request. Both call sites that
// flatten a template's attributes share this one renderer and this one order.

func TestAttributeStringRendersNumbersWithoutScientificNotation(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"a price", float64(1250000), "1250000"},
		{"a fuel tank", float64(15.8), "15.8"},
		{"a door count", float64(4), "4"},
		{"a large round number", float64(1e21), "1000000000000000000000"},
		{"a small fraction", float64(0.25), "0.25"},
		{"a string", "AWD", "AWD"},
		{"a bool", true, "true"},
		{"an int", 7, "7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AttributeString(tt.value))
			assert.NotContains(t, AttributeString(tt.value), "e+", "a value rendered in scientific notation is not the number a client asked for")
		})
	}
}

// The order has to be the same on every call for the same template, or the
// attribute array shuffles between two requests for one style.
func TestSortedAttributeNamesIsStable(t *testing.T) {
	attributes := map[string]any{
		"number_of_doors":        float64(4),
		"base_msrp":              float64(1250000),
		"driven_wheels":          "AWD",
		"fuel_tank_capacity_gal": float64(15.8),
		"powertrain_type":        "ICE",
	}

	want := []string{"base_msrp", "driven_wheels", "fuel_tank_capacity_gal", "number_of_doors", "powertrain_type"}
	for i := 0; i < 20; i++ {
		assert.Equal(t, want, SortedAttributeNames(attributes))
	}
}

func TestSortedAttributeNamesHandlesNoAttributes(t *testing.T) {
	assert.Empty(t, SortedAttributeNames(nil))
	assert.Empty(t, SortedAttributeNames(map[string]any{}))
}
