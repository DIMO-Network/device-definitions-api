package gateways

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func Test_isBodyStyleCode(t *testing.T) {
	cases := map[string]bool{
		"4D":    true,
		"5D":    true,
		"2D":    true,
		"4DR":   true,
		"5HB":   true,
		"123":   true,
		"Crown": false,
		"CAMRY": false,
		"":      false,
		"Q7":    false, // letter+digit, not digit-leading
	}
	for in, want := range cases {
		assert.Equalf(t, want, isBodyStyleCode(in), "input %q", in)
	}
}

func Test_pickModelCandidate(t *testing.T) {
	assert.Equal(t, "Crown", pickModelCandidate("Crown"))
	assert.Equal(t, "CAMRY", pickModelCandidate("CAMRY/HYBRID"))
	assert.Equal(t, "Crown", pickModelCandidate("4D/Crown"))
	assert.Equal(t, "", pickModelCandidate("4D"))
	assert.Equal(t, "", pickModelCandidate(""))
}

func Test_extractModelName_rejectsBodyStyle(t *testing.T) {
	// simulates prior bug: Model Name column held "4D"; series name lives in additional info.
	payload := `{
		"data": {
			"model_original_epc_list": [{
				"CarAttributes": [
					{"Col_name": "Model Name", "Col_value": "4D"},
					{"Col_name": "Additional Vehicle Infomation", "Col_value": "Crown Hybrid"}
				]
			}]
		}
	}`
	got := extractModelName(gjson.Parse(payload))
	assert.Equal(t, "Crown", got)
}

func Test_extractModelName_prefersModelNameWhenValid(t *testing.T) {
	payload := `{
		"data": {
			"model_original_epc_list": [{
				"CarAttributes": [
					{"Col_name": "Model Name", "Col_value": "CAMRY/HYBRID"},
					{"Col_name": "Additional Vehicle Infomation", "Col_value": "LHD CHI"}
				]
			}]
		}
	}`
	assert.Equal(t, "CAMRY", extractModelName(gjson.Parse(payload)))
}

func Test_extractModelName_fallsThroughColumnVariants(t *testing.T) {
	payload := `{
		"data": {
			"model_original_epc_list": [{
				"CarAttributes": [
					{"Col_name": "Model name", "Col_value": "Crown"}
				]
			}]
		}
	}`
	assert.Equal(t, "Crown", extractModelName(gjson.Parse(payload)))
}

func Test_extractModelName_chineseFallback(t *testing.T) {
	payload := `{
		"data": {
			"model_original_epc_list": [{
				"CarAttributes": [
					{"Col_name": "车型", "Col_value": "Crown"}
				]
			}]
		}
	}`
	assert.Equal(t, "Crown", extractModelName(gjson.Parse(payload)))
}

func Test_extractModelName_empty(t *testing.T) {
	assert.Equal(t, "", extractModelName(gjson.Parse(`{"data":{}}`)))
}
