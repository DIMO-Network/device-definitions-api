package gateways

import (
	_ "embed" // import the embed package
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

//go:embed test_vincario_resp.json
var testVincarioVINResp []byte

func Test_vincarioAPIService_DecodeVIN(t *testing.T) {
	const testVIN = "WAUZZZ4M0KD018683"
	const baseUrl = "http://local"

	vincarioSvc := NewVincarioAPIService(&config.Settings{VincarioAPIKey: "xxx", VincarioAPISecret: "XXX", VincarioAPIURL: baseUrl}, dbtest.Logger())
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	path := vincarioPathBuilder(testVIN, "decode", "xxx", "XXX")

	httpmock.RegisterResponder(http.MethodGet, baseUrl+path, httpmock.NewStringResponder(200, string(testVincarioVINResp)))

	resp, err := vincarioSvc.DecodeVIN(testVIN)

	require.NoError(t, err)
	assert.Equal(t, testVIN, resp.VIN)
	// todo fill in

}
