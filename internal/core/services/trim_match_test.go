package services

import (
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func camry() *coremodels.Template {
	return &coremodels.Template{
		ID: "toyota_camry_2020", DeviceType: "vehicle",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota", TokenID: 131},
		Model:        "Camry", Year: 2020, Version: 3,
		Attributes: map[string]any{"number_of_doors": float64(4)},
		Trims: []coremodels.Trim{
			{Name: "LE", Selectors: coremodels.TrimSelectors{ManufacturerCode: []string{"2532"}},
				Attributes: map[string]any{"powertrain_type": "ICE", "fuel_tank_capacity_gal": 16.0, "driven_wheels": "FWD"}},
			{Name: "Hybrid LE", Selectors: coremodels.TrimSelectors{ManufacturerCode: []string{"2559"}},
				Attributes: map[string]any{"powertrain_type": "HEV", "fuel_tank_capacity_gal": 13.2, "driven_wheels": "FWD"}},
		},
	}
}

func TestMatchTrim_ExactOnManufacturerCode(t *testing.T) {
	r := MatchTrim(camry(), MatchSignals{ManufacturerCode: "2559"})
	assert.Equal(t, MatchExact, r.Quality)
	assert.Equal(t, "Hybrid LE", r.Trim)
	assert.Equal(t, []string{"manufacturerCode"}, r.MatchedBy)
	assert.Equal(t, "HEV", r.Attributes["powertrain_type"])
	assert.Equal(t, 13.2, r.Attributes["fuel_tank_capacity_gal"])
	// Template-level attributes merge in.
	assert.Equal(t, float64(4), r.Attributes["number_of_doors"])
	assert.Equal(t, 3, r.TemplateVersion)
}

func TestMatchTrim_TheBlendedCamryIsNoLongerPossible(t *testing.T) {
	// The production record declares ICE while carrying the hybrid's 13.2 gal
	// tank. Whichever trim matches, powertrain and tank must come from the
	// SAME trim.
	for _, tc := range []struct {
		code, pt string
		tank     float64
	}{
		{"2532", "ICE", 16.0},
		{"2559", "HEV", 13.2},
	} {
		r := MatchTrim(camry(), MatchSignals{ManufacturerCode: tc.code})
		assert.Equal(t, tc.pt, r.Attributes["powertrain_type"])
		assert.Equal(t, tc.tank, r.Attributes["fuel_tank_capacity_gal"])
	}
}

func TestMatchTrim_AmbiguousEmitsOnlyAgreedAttributes(t *testing.T) {
	tmpl := camry()
	// Both trims claim the same code: a template defect, not a decode failure.
	tmpl.Trims[1].Selectors.ManufacturerCode = []string{"2532"}
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	assert.Equal(t, MatchAmbiguous, r.Quality)
	assert.ElementsMatch(t, []string{"LE", "Hybrid LE"}, r.Candidates)
	assert.Empty(t, r.Trim)
	// They disagree on powertrain and tank, agree on drive.
	assert.NotContains(t, r.Attributes, "powertrain_type")
	assert.NotContains(t, r.Attributes, "fuel_tank_capacity_gal")
	assert.Equal(t, "FWD", r.Attributes["driven_wheels"])
}

func TestMatchTrim_ModelOnlyWhenNothingMatches(t *testing.T) {
	r := MatchTrim(camry(), MatchSignals{ManufacturerCode: "9999"})
	assert.Equal(t, MatchModelOnly, r.Quality)
	assert.Empty(t, r.Trim)
	// Only template-level attributes survive; no trim's values leak in.
	assert.Equal(t, float64(4), r.Attributes["number_of_doors"])
	assert.NotContains(t, r.Attributes, "powertrain_type")
}

func TestMatchTrim_SingleTrimTemplateMatchesWithoutSelectors(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = tmpl.Trims[:1]
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{}
	r := MatchTrim(tmpl, MatchSignals{})
	assert.Equal(t, MatchExact, r.Quality)
	assert.Equal(t, "LE", r.Trim)
}

func TestMatchTrim_StyleNameIsCaseInsensitive(t *testing.T) {
	tmpl := camry()
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{StyleName: []string{"LE 4dr Sedan"}}
	r := MatchTrim(tmpl, MatchSignals{StyleName: "le 4dr sedan"})
	assert.Equal(t, MatchExact, r.Quality)
	assert.Equal(t, "LE", r.Trim)
}

func TestMatchTrim_TrimOverridesTemplateHardwareTemplateID(t *testing.T) {
	tmpl := camry()
	tmpl.HardwareTemplateID = "130"
	tmpl.Trims[1].HardwareTemplateID = "115"
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2559"})
	assert.Equal(t, "115", r.HardwareTemplateID)
	r2 := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	assert.Equal(t, "130", r2.HardwareTemplateID)
}

func TestMatchTrim_EmptyTemplateIsModelOnlyNotAPanic(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = nil
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	assert.Equal(t, MatchModelOnly, r.Quality)
	require.NotNil(t, r.Attributes)
}

func TestMatchTrim_HardwareTemplateIDStaysTemplatesWhenAmbiguous(t *testing.T) {
	tmpl := camry()
	tmpl.HardwareTemplateID = "130"
	tmpl.Trims[0].HardwareTemplateID = "115"
	tmpl.Trims[1].HardwareTemplateID = "999"
	// Both trims claim the same code: a template defect, not a decode failure.
	tmpl.Trims[1].Selectors.ManufacturerCode = []string{"2532"}
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	require.Equal(t, MatchAmbiguous, r.Quality)
	// Neither candidate's HardwareTemplateID wins -- that would still be
	// picking one, in the field that decides what hardware DIMO ships.
	assert.Equal(t, "130", r.HardwareTemplateID)
}

func TestMatchTrim_HardwareTemplateIDStaysTemplatesWhenModelOnly(t *testing.T) {
	tmpl := camry()
	tmpl.HardwareTemplateID = "130"
	tmpl.Trims[0].HardwareTemplateID = "115"
	tmpl.Trims[1].HardwareTemplateID = "999"
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "9999"})
	require.Equal(t, MatchModelOnly, r.Quality)
	assert.Equal(t, "130", r.HardwareTemplateID)
}

func TestMatchTrim_VINPatternIsAnchoredToTheWholeVIN(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = tmpl.Trims[:1]
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{VINPattern: "4T1B11HK.{9}"}

	whole := MatchTrim(tmpl, MatchSignals{VIN: "4T1B11HK5LU123456"})
	assert.Equal(t, MatchExact, whole.Quality)

	// The same pattern is a bare substring of a longer VIN. Unanchored,
	// regexp.MatchString would find it anywhere in the string; anchored, it
	// must not match unless the pattern accounts for the whole VIN.
	substring := MatchTrim(tmpl, MatchSignals{VIN: "XX4T1B11HK5LU123456XX"})
	assert.Equal(t, MatchModelOnly, substring.Quality)
}

func TestMatchTrim_InvalidVINPatternDoesNotMatchOrPanic(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = tmpl.Trims[:1]
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{VINPattern: "("}
	assert.NotPanics(t, func() {
		r := MatchTrim(tmpl, MatchSignals{VIN: "4T1B11HK5LU123456"})
		assert.Equal(t, MatchModelOnly, r.Quality)
	})
}

func TestMatchTrim_EmptyVINSignalDoesNotMatchVINPattern(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = tmpl.Trims[:1]
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{VINPattern: "4T1B11HK.{9}"}
	r := MatchTrim(tmpl, MatchSignals{})
	assert.Equal(t, MatchModelOnly, r.Quality)
}
