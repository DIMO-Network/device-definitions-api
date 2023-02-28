package models

import (
	"github.com/volatiletech/null/v8"
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
