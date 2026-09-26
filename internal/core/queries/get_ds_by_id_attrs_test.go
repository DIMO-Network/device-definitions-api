package queries

import (
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GET /device-styles/{id} serves the attributes of the definition the style
// belongs to. A template carries at the top level only what every trim agrees
// on, so copying those alone drops every trim-varying attribute and reports no
// powertrain for any model sold as combustion, hybrid and plug-in hybrid --
// after which the name heuristic misses "Prime SE" and the code asserts ICE
// for a plug-in hybrid. The style's name is exactly the signal
// services.MatchTrim consumes, so the trim is resolved and its attributes
// served.
//
// These exercise the resolution itself, which is pure: no Postgres, no
// catalog. Only the handler's device_styles read needs CI.

// rav4 is the shape the finding describes: one model sold with three
// powertrains, so powertrain_type lives on each trim and never at the top.
func rav4() *coremodels.Template {
	return &coremodels.Template{
		ID:           "toyota_rav4_2021",
		DeviceType:   "vehicle",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota", TokenID: 131},
		Model:        "RAV4",
		Year:         2021,
		Attributes: map[string]any{
			"number_of_doors": float64(4),
			"driven_wheels":   "AWD",
		},
		Trims: []coremodels.Trim{
			{
				Name:       "LE",
				Selectors:  coremodels.TrimSelectors{StyleName: []string{"LE"}},
				Attributes: map[string]any{common.PowerTrainType: "ICE", "mpg": float64(30), "base_msrp": float64(1250000)},
			},
			{
				Name:       "XSE Hybrid",
				Selectors:  coremodels.TrimSelectors{StyleName: []string{"XSE Hybrid"}},
				Attributes: map[string]any{common.PowerTrainType: "HEV", "mpg": float64(40)},
			},
			{
				Name:       "Prime SE",
				Selectors:  coremodels.TrimSelectors{StyleName: []string{"Prime SE"}},
				Attributes: map[string]any{common.PowerTrainType: "PHEV", "mpg": float64(94), "fuel_tank_capacity_gal": float64(14.5)},
			},
		},
		Version: 3,
	}
}

func attributeByName(t *testing.T, attrs []coremodels.DeviceTypeAttributeEditor, name string) string {
	t.Helper()
	for _, a := range attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func hasAttribute(attrs []coremodels.DeviceTypeAttributeEditor, name string) bool {
	for _, a := range attrs {
		if a.Name == name {
			return true
		}
	}
	return false
}

// The finding itself: a plug-in hybrid whose name no heuristic recognises.
func TestStyleAttributesResolveThePlugInHybridTrim(t *testing.T) {
	attrs := styleDeviceAttributes(rav4(), "Prime SE", nil)

	assert.Equal(t, "PHEV", attributeByName(t, attrs, common.PowerTrainType),
		"the matched trim says PHEV; asserting ICE here is the defect")
	assert.Equal(t, "94", attributeByName(t, attrs, "mpg"), "a trim-varying attribute must be served")
	assert.Equal(t, "14.5", attributeByName(t, attrs, "fuel_tank_capacity_gal"))
	// The template's own attributes are still there.
	assert.Equal(t, "4", attributeByName(t, attrs, "number_of_doors"))
	assert.Equal(t, "AWD", attributeByName(t, attrs, "driven_wheels"))
}

// The trim the heuristic would have got right must still come out right, and
// with its own attributes rather than another trim's.
func TestStyleAttributesResolveTheHybridTrim(t *testing.T) {
	attrs := styleDeviceAttributes(rav4(), "XSE Hybrid", nil)

	assert.Equal(t, "HEV", attributeByName(t, attrs, common.PowerTrainType))
	assert.Equal(t, "40", attributeByName(t, attrs, "mpg"))
	assert.False(t, hasAttribute(attrs, "fuel_tank_capacity_gal"), "attributes of a trim that did not match must not leak in")
}

func TestStyleAttributesResolveTheCombustionTrim(t *testing.T) {
	attrs := styleDeviceAttributes(rav4(), "LE", nil)

	assert.Equal(t, "ICE", attributeByName(t, attrs, common.PowerTrainType))
	assert.Equal(t, "30", attributeByName(t, attrs, "mpg"))
}

// What the style itself recorded at decode time wins: it is the one value that
// was observed for this style rather than inferred from a name.
func TestStyleAttributesPreferThePowertrainOnTheStyleRow(t *testing.T) {
	attrs := styleDeviceAttributes(rav4(), "Prime SE", []byte(`{"powertrain_type":"BEV"}`))

	assert.Equal(t, "BEV", attributeByName(t, attrs, common.PowerTrainType))
}

// No trim matched, so nothing carries a powertrain -- but the name does. The
// heuristic still answers, as it always has.
func TestStyleAttributesFallBackToTheNameHeuristicWhenNoTrimMatches(t *testing.T) {
	attrs := styleDeviceAttributes(rav4(), "Limited Plug-In", nil)

	assert.Equal(t, models.PowertrainPHEV, attributeByName(t, attrs, common.PowerTrainType))
	// No trim was chosen, so no trim's attributes are served.
	assert.False(t, hasAttribute(attrs, "mpg"), "an unmatched style must not be given some trim's attributes")
	assert.Equal(t, "4", attributeByName(t, attrs, "number_of_doors"))
}

// A model where every trim agrees on the powertrain carries it at the top
// level. An unmatched style must take that, not the ICE default.
func TestStyleAttributesUseTheTemplatePowertrainWhenNoTrimMatches(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:         "tesla_model-3_2022",
		Model:      "Model 3",
		Year:       2022,
		Attributes: map[string]any{common.PowerTrainType: "BEV"},
		Trims: []coremodels.Trim{
			{Name: "Long Range", Selectors: coremodels.TrimSelectors{StyleName: []string{"Long Range"}}},
		},
	}

	attrs := styleDeviceAttributes(tmpl, "Some Style We Do Not Know", nil)

	assert.Equal(t, "BEV", attributeByName(t, attrs, common.PowerTrainType),
		"the template says BEV for every trim; defaulting to ICE would contradict it")
}

// Two trims match, so no trim is chosen: only what every candidate agrees on
// is served, and the powertrain they agree on is not a guess.
func TestStyleAttributesServeOnlyWhatAmbiguousCandidatesAgreeOn(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:    "toyota_camry_2026",
		Model: "Camry",
		Year:  2026,
		Trims: []coremodels.Trim{
			{Name: "LE", Selectors: coremodels.TrimSelectors{StyleName: []string{"LE"}}, Attributes: map[string]any{common.PowerTrainType: "HEV", "mpg": float64(44)}},
			{Name: "LE FWD", Selectors: coremodels.TrimSelectors{StyleName: []string{"LE"}}, Attributes: map[string]any{common.PowerTrainType: "HEV", "mpg": float64(51)}},
		},
	}

	attrs := styleDeviceAttributes(tmpl, "LE", nil)

	assert.Equal(t, "HEV", attributeByName(t, attrs, common.PowerTrainType), "both candidates say HEV")
	assert.False(t, hasAttribute(attrs, "mpg"), "the candidates disagree on mpg, so neither answer may be served")
}

// The ICE default is the last resort and stays: nothing anywhere -- the style
// row, the matched trim, the template, the name -- says anything about the
// powertrain.
func TestStyleAttributesDefaultToICEOnlyWhenThePowertrainIsUnknown(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:         "ford_escape_2020",
		Model:      "Escape",
		Year:       2020,
		Attributes: map[string]any{"number_of_doors": float64(4)},
	}

	attrs := styleDeviceAttributes(tmpl, "SE", nil)

	assert.Equal(t, models.PowertrainICE, attributeByName(t, attrs, common.PowerTrainType))
}

// A style whose name cannot narrow anything still gets an answer, and the
// resolution never panics on a template with no trims at all.
func TestStyleAttributesHandleATemplateWithNoTrims(t *testing.T) {
	tmpl := &coremodels.Template{
		ID:         "toyota_camry_2026",
		Model:      "Camry",
		Year:       2026,
		Attributes: map[string]any{common.PowerTrainType: "ICE", "number_of_doors": float64(4)},
	}

	attrs := styleDeviceAttributes(tmpl, "", nil)

	require.NotEmpty(t, attrs)
	assert.Equal(t, "ICE", attributeByName(t, attrs, common.PowerTrainType))
}

// A template attribute decodes as a float64, so fmt.Sprint rendered a price of
// 1250000 as "1.25e+06" -- while GetDeviceDefinitionByID served "1250000" for
// the same definition, because the gateway already had a renderer written for
// this. Both now share it.
func TestStyleAttributesRenderNumbersWithoutScientificNotation(t *testing.T) {
	attrs := styleDeviceAttributes(rav4(), "LE", nil)

	assert.Equal(t, "1250000", attributeByName(t, attrs, "base_msrp"))
	assert.Equal(t, "4", attributeByName(t, attrs, "number_of_doors"))
	assert.Equal(t, "30", attributeByName(t, attrs, "mpg"))
	for _, a := range attrs {
		assert.NotContains(t, a.Value, "e+", "%s must not be served in scientific notation", a.Name)
	}
}

// The attributes came out of a Go map, so the array was in a different order
// on every request for the same style. The sibling definition path sorts.
func TestStyleAttributesAreInAStableOrder(t *testing.T) {
	want := namesOf(styleDeviceAttributes(rav4(), "Prime SE", nil))
	assert.Equal(t, []string{"driven_wheels", "fuel_tank_capacity_gal", "mpg", "number_of_doors", "powertrain_type"}, want)
	for i := 0; i < 20; i++ {
		assert.Equal(t, want, namesOf(styleDeviceAttributes(rav4(), "Prime SE", nil)),
			"the attribute array must not shuffle between requests for one style")
	}
}

func namesOf(attrs []coremodels.DeviceTypeAttributeEditor) []string {
	names := make([]string, 0, len(attrs))
	for _, a := range attrs {
		names = append(names, a.Name)
	}
	return names
}
