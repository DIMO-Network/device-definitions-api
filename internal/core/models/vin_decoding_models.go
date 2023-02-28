package models

import (
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/volatiletech/null/v8"
	"strconv"
	"strings"
)

type DecodeProviderEnum string

const (
	DrivlyProvider   DecodeProviderEnum = "drivly"
	VincarioProvider DecodeProviderEnum = "vincario"
	AllProviders     DecodeProviderEnum = ""
)

type VINDecodingInfoData struct {
	VIN        string
	Make       string
	Model      string
	SubModel   string
	Year       string
	StyleName  string
	Source     DecodeProviderEnum
	ExternalID string
	MetaData   null.JSON
}

func (v *VINDecodingInfoData) LoadFromVincario(info *gateways.VincarioInfoResponse) {
	v.VIN = info.VIN
	v.Year = strconv.Itoa(info.ModelYear)
	v.Make = strings.TrimSpace(info.Make)
	v.Model = strings.TrimSpace(info.Model)
	v.Source = VincarioProvider
	v.ExternalID = strconv.Itoa(info.VehicleID)
	v.StyleName = info.GetStyle()
	v.SubModel = info.GetSubModel()
}

func (v *VINDecodingInfoData) LoadFromDrivly(info *gateways.DrivlyVINResponse) {
	v.VIN = info.Vin
	v.Year = info.Year
	v.Make = info.Make
	v.Model = info.Model
	v.StyleName = buildDrivlyStyleName(info)
	v.ExternalID = info.GetExternalID()
	v.Source = DrivlyProvider
}

func buildDrivlyStyleName(vinInfo *gateways.DrivlyVINResponse) string {
	return strings.TrimSpace(vinInfo.Trim + " " + vinInfo.SubModel)
}
