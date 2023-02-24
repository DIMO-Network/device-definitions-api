package models

import "github.com/volatiletech/null/v8"

type VINDecodingInfoData struct {
	VIN        string
	Make       string
	Model      string
	SubModel   string
	Year       string
	StyleName  string
	Source     string
	ExternalID string
	MetaData   null.JSON
}
