package gateways

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"time"
)

//go:generate mockgen -source vincario_api_service.go -destination mocks/vincario_api_service_mock.go -package mocks
type VincarioAPIService interface {
	DecodeVIN(vin string) (any, error)
}

type vincarioAPIService struct {
	settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
}

func NewVincarioAPIService(settings *config.Settings) VincarioAPIService {
	if settings.VincarioAPIURL == "" || settings.VincarioAPISecret == "" {
		panic("Vincario configuration not set")
	}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.VincarioAPIURL, "", 10*time.Second, nil, false)

	return &vincarioAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
	}
}

func (va *vincarioAPIService) DecodeVIN(vin string) (any, error) {
	id := "decode"

	s := vin + "|" + id + "|" + va.settings.VincarioAPIKey + "|" + va.settings.VincarioAPISecret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest("/"+va.settings.VincarioAPIKey+"/"+controlSum+"/"+id+"/"+vin+".json", "GET", nil)
	if err != nil {
		return nil, err
	}

	// decode JSON from response body
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return data, nil
}
