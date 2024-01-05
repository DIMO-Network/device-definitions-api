package models

import (
	"github.com/volatiletech/null/v8"
)

type DecodeProviderEnum string

const (
	DrivlyProvider   DecodeProviderEnum = "drivly"
	VincarioProvider DecodeProviderEnum = "vincario"
	AutoIsoProvider  DecodeProviderEnum = "autoiso"
	AllProviders     DecodeProviderEnum = ""
)

type VINDecodingInfoData struct {
	VIN        string
	Make       string
	Model      string
	SubModel   string
	Year       int32
	StyleName  string
	Source     DecodeProviderEnum
	ExternalID string
	MetaData   null.JSON
	Raw        []byte
	FuelType   string
}
