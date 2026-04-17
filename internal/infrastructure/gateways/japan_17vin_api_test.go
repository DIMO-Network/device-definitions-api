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
	assert.Equal(t, "Crown", pickModelCandidate("Crown", ""))
	assert.Equal(t, "CAMRY", pickModelCandidate("CAMRY/HYBRID", ""))
	assert.Equal(t, "Crown", pickModelCandidate("4D/Crown", ""))
	assert.Equal(t, "", pickModelCandidate("4D", ""))
	assert.Equal(t, "", pickModelCandidate("", ""))
	// hint picks the platform variant actually built
	assert.Equal(t, "VOXY", pickModelCandidate("NOAH/VOXY", "VOXY 07S  HTWC CBU"))
	assert.Equal(t, "VOXY", pickModelCandidate("NOAH/VOXY/ESQUIRE", "VOXY 07S  HTWVMCBU"))
	// hint irrelevant when it doesn't match any candidate: first wins
	assert.Equal(t, "CROWN", pickModelCandidate("CROWN/HYBRID", "4D   HTWC"))
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

// Data-driven regression suite using real 17vin payloads. Each case pairs the
// full `data.model_original_epc_list` attributes with the model name we expect
// the adapter to surface (compare against the DD slug historically indexed).
func Test_extractModelName_realPayloads(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    string
	}{
		{
			name: "JTDZN3EU8HJ060118 Prius v",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"PRIUS V"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"05S  USA"}
			]}]}}`,
			want: "PRIUS V",
		},
		{
			name: "TRJ150-0081549 Land Cruiser Prado",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"LAND CRUISER PRADO"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"5D   07S"}
			]}]}}`,
			want: "LAND CRUISER PRADO",
		},
		{
			name: "GRX120-3043102 Mark X",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"MARK X"}
			]}]}}`,
			want: "MARK X",
		},
		{
			name: "JTEBU29J940026005 Land Cruiser Prado",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"LAND CRUISER PRADO"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"LHD  5D"}
			]}]}}`,
			want: "LAND CRUISER PRADO",
		},
		{
			name: "ZWR90-8000186 Voxy (NOAH/VOXY disambiguated by addl info)",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"NOAH/VOXY"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"VOXY 07S  HTWC CBU"}
			]}]}}`,
			want: "VOXY",
		},
		{
			name: "3TYLC5LN5ST036165 Tacoma",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"TACOMA"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"DCB"}
			]}]}}`,
			want: "TACOMA",
		},
		{
			name: "A210A-0012622 Raize",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"RAIZE"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"5D   05S"}
			]}]}}`,
			want: "RAIZE",
		},
		{
			name: "AGH30-0397617 Alphard (previously ALPHD07S bug)",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"ALPHARD/VELLFIRE/HV"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"ALPHD07S"}
			]}]}}`,
			want: "ALPHARD",
		},
		{
			name: "ZVW60-4006778 Prius",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"PRIUS"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"5D"}
			]}]}}`,
			want: "PRIUS",
		},
		{
			name: "ZRR80-0413374 Voxy (NOAH/VOXY/ESQUIRE disambiguated)",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"NOAH/VOXY/ESQUIRE"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"VOXY 07S  HTWVMCBU"}
			]}]}}`,
			want: "VOXY",
		},
		{
			name: "SJNFAAJ11U2132164 Qashqai UK Make",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"QASHQAI UK MAKE"}
			]}]}}`,
			want: "QASHQAI UK MAKE",
		},
		{
			name: "NHP10-6705283 Aqua",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"AQUA"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"5D"}
			]}]}}`,
			want: "AQUA",
		},
		{
			name: "NHP170-7078636 Sienta",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"SIENTA"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"07S  HTWC"}
			]}]}}`,
			want: "SIENTA",
		},
		{
			name: "7MUAAABG3PV055668 Corolla Cross",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"COROLLA CROSS"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"LHD"}
			]}]}}`,
			want: "COROLLA CROSS",
		},
		{
			name: "SUNFAAZE1U0009011 Leaf UK Make",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"LEAF UK MAKE"}
			]}]}}`,
			want: "LEAF UK MAKE",
		},
		{
			name: "GWS214-6014148 Crown (slash with HYBRID variant)",
			payload: `{"data":{"model_original_epc_list":[{"CarAttributes":[
				{"Col_name":"Model Name","Col_value":"CROWN/HYBRID"},
				{"Col_name":"Additional Vehicle Infomation","Col_value":"4D   HTWC"}
			]}]}}`,
			want: "CROWN",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, extractModelName(gjson.Parse(tc.payload)))
		})
	}
}

// Regression: real 17vin response for GWS214-6014148 (Toyota Crown Hybrid).
// Previously yielded "4D" from the "Additional Vehicle Infomation" column and
// got minted on-chain as toyota_4d_2017.
func Test_extractModelName_gws214Crown(t *testing.T) {
	payload := `{
		"data": {
			"epc": "toyota",
			"model_year_from_vin": "2017",
			"model_original_epc_list": [{
				"CarAttributes": [
					{"Col_name": "车型", "Col_value": "CROWN/HYBRID"},
					{"Col_name": "Model Name", "Col_value": "CROWN/HYBRID"},
					{"Col_name": "车型代码", "Col_value": "GWS214-AEXZB"},
					{"Col_name": "Model Code", "Col_value": "GWS214-AEXZB"},
					{"Col_name": "车身", "Col_value": "SED"},
					{"Col_name": "Body", "Col_value": "SED"},
					{"Col_name": "Engine Code", "Col_value": "2GRFXE"},
					{"Col_name": "Additional Vehicle Infomation", "Col_value": "4D   HTWC"}
				]
			}]
		}
	}`
	assert.Equal(t, "CROWN", extractModelName(gjson.Parse(payload)))
}
