package gateways

import (
	_ "embed" // import the embed package
	"net/http"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_vincario_resp.json
var testVincarioVINResp []byte

func Test_vincarioAPIService_DecodeVIN(t *testing.T) {
	const testVIN = "WAUZZZ4M0KD018683"
	const baseURL = "http://local"

	vincarioSvc := NewVincarioAPIService(&config.Settings{VincarioAPIKey: "xxx", VincarioAPISecret: "XXX", VincarioAPIURL: baseURL}, dbtest.Logger())
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	path := vincarioPathBuilder(testVIN, "decode", "xxx", "XXX")

	httpmock.RegisterResponder(http.MethodGet, baseURL+path, httpmock.NewStringResponder(200, string(testVincarioVINResp)))

	resp, err := vincarioSvc.DecodeVIN(testVIN)

	require.NoError(t, err)
	assert.Equal(t, testVIN, resp.VIN)
	assert.Equal(t, "Q7", resp.Model)
	assert.Equal(t, "Audi", resp.Make)
	assert.Equal(t, "Wagon", resp.Body)
	assert.Equal(t, "Diesel", resp.FuelType)
	assert.Equal(t, 2019, resp.ModelYear)
	assert.Equal(t, "II (2015-)", resp.Series)
	assert.Equal(t, `4-Stroke / 6 / V-T-DI`, resp.EngineType)
	assert.Equal(t, 2967, resp.EngineDisplacement)
	assert.Equal(t, "DHX", resp.EngineCode)
	assert.Equal(t, "4x4 - Four-wheel drive", resp.Drive)
	assert.Equal(t, 191, resp.VehicleID)
	assert.Equal(t, "Automatic", resp.Transmission)
	assert.Equal(t, 8, resp.NumberOfGears)
	// style: {FuelType} {enginetype} {transmission} {numberofgears} speed
	s := resp.GetStyle()
	assert.Equal(t, "Diesel 4-Stroke / 6 / V-T-DI Automatic 8-speed", s)
	//assert.Equal(t, 191, vi)
}

func Test_vincarioAPIService_DecodeVIN_MissingProperties(t *testing.T) {
	const testVIN = "WAUZZZ4M0KD018683"
	const baseURL = "http://local"

	vincarioSvc := NewVincarioAPIService(&config.Settings{VincarioAPIKey: "xxx", VincarioAPISecret: "XXX", VincarioAPIURL: baseURL}, dbtest.Logger())
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	path := vincarioPathBuilder(testVIN, "decode", "xxx", "XXX")

	httpmock.RegisterResponder(http.MethodGet, baseURL+path, httpmock.NewStringResponder(200, string(testVincarioVINResp)))

	resp, err := vincarioSvc.DecodeVIN(testVIN)

	require.NoError(t, err)
	assert.Equal(t, testVIN, resp.VIN)
	assert.Equal(t, "Q7", resp.Model)
	assert.Equal(t, "Audi", resp.Make)
	assert.Equal(t, "Wagon", resp.Body)
	assert.Equal(t, "Diesel", resp.FuelType)
	assert.Equal(t, 2019, resp.ModelYear)
	assert.Equal(t, "II (2015-)", resp.Series)
	assert.Equal(t, `4-Stroke / 6 / V-T-DI`, resp.EngineType)
	assert.Equal(t, 2967, resp.EngineDisplacement)
	assert.Equal(t, "DHX", resp.EngineCode)
	assert.Equal(t, "4x4 - Four-wheel drive", resp.Drive)
	assert.Equal(t, 191, resp.VehicleID)
	assert.Equal(t, "Automatic", resp.Transmission)
	assert.Equal(t, 8, resp.NumberOfGears)
	// style: {FuelType} {enginetype} {transmission} {numberofgears} speed
	s := resp.GetStyle()
	assert.Equal(t, "Diesel 4-Stroke / 6 / V-T-DI Automatic 8-speed", s)
	//assert.Equal(t, 191, vi)
}
