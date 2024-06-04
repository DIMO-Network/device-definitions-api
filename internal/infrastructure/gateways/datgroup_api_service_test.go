package gateways

import (
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"

	"github.com/stretchr/testify/assert"
)

//go:embed test_datgoup_resp1.xml
var testDatgroupXML1 []byte

func Test_parseXML(t *testing.T) {
	logger := dbtest.Logger()

	response, err := parseXML(logger, string(testDatgroupXML1), "WVWZZZE1ZP8005474")

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
	// Series Equipment
	assert.Equal(t, "38937", response.SeriesEquipment[0].DatEquipmentId)
	assert.Equal(t, "GM1", response.SeriesEquipment[0].ManufacturerEquipmentId)
	assert.Equal(t, "Electronic engine sound", response.SeriesEquipment[0].ManufacturerDescription)
	assert.Equal(t, "acoustic pedestrian protection, external sound (e-sound)", response.SeriesEquipment[0].Description)
	// SpecialEquipment
	assert.Equal(t, "11166", response.SpecialEquipment[0].DatEquipmentId)
	assert.Equal(t, "C2A1", response.SpecialEquipment[0].ManufacturerEquipmentId)
	assert.Equal(t, "Moonstone Gray/Black", response.SpecialEquipment[0].ManufacturerDescription)
	assert.Equal(t, "custom paintwork Moonstone grey single-tone", response.SpecialEquipment[0].Description)
	// DATECodeEquipment
	assert.Equal(t, "97194", response.DATECodeEquipment[0].DatEquipmentId)
	assert.Equal(t, "electric motor 150 kW (cont. 70 kW)", response.DATECodeEquipment[0].Description)
	// VINEquipment
	assert.Equal(t, "0FY", response.VINEquipment[0].ManufacturerEquipmentId)
	assert.Equal(t, "Dresden manufacturing sequence", response.VINEquipment[0].ManufacturerDescription)
}
