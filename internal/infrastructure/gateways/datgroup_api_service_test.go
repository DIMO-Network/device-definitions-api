package gateways

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed test_datgoup_resp1.xml
var testDatgroupXml1 []byte

func Test_parseXML(t *testing.T) {
	response, err := parseXML(string(testDatgroupXml1), "WVWZZZE1ZP8005474")

	assert.NoError(t, err)
	assert.Equal(t, "019051530110001", response.DatECode)
	assert.Equal(t, "ID.3 (E11)(06.2020->2023)", response.BaseModelName)
	assert.Equal(t, "ID.3", response.MainTypeGroupName)
	assert.Equal(t, "Volkswagen", response.ManufacturerName)
	assert.Equal(t, "ID.3 Pro Performance", response.SalesDescription)
	assert.Equal(t, "Pro Performance 150 kW", response.SubModelName)
	assert.Equal(t, "Passenger car, SUV, small van", response.VehicleTypeName)
	assert.Equal(t, 0, response.VinAccuracy)
	assert.Equal(t, 2020, response.YearLow)
	assert.Equal(t, 2023, response.YearHigh)
	assert.Equal(t, 2023, response.Year)

}
