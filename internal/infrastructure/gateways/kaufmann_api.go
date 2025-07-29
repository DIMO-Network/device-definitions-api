package gateways

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
)

//go:generate mockgen -source kaufmann_api.go -destination mocks/kaufmann_api_mock.go -package mocks

type ElevaConfig struct {
	Username string
	Password string
}

type ElevaAPI interface {
	GetVINInfo(vin string) (*coremodels.ElevaVINResponse, error)
}

// ElevaAPI is used to call Kaufmann for VIN decoding in Chile
type elevaAPI struct {
	client       *http.Client
	config       ElevaConfig
	accessToken  string
	tokenExpires time.Time
}

func NewElevaAPI(settings *config.Settings) ElevaAPI {
	return &elevaAPI{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		config: ElevaConfig{
			Username: settings.ElevaUsername,
			Password: settings.ElevaPassword,
		},
	}
}

// how long does the access token last
const tokenExpiration = 15 * time.Minute // assumed

func (e *elevaAPI) getAccessToken() error {
	if e.accessToken != "" && time.Now().Before(e.tokenExpires) {
		return nil
	}

	loginURL := "https://api-prd-js.eleva-services.com/v4/auth/login"

	payload := map[string]string{
		"username": e.config.Username,
		"password": e.config.Password,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://portalprepago.elevapp.io")
	req.Header.Set("referer", "https://portalprepago.elevapp.io/")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("performing login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("auth failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	if result.Data.AccessToken == "" {
		return fmt.Errorf("no access token returned")
	}

	e.accessToken = result.Data.AccessToken
	e.tokenExpires = time.Now().Add(tokenExpiration)
	return nil
}

func (e *elevaAPI) GetVINInfo(plateOrVIN string) (*coremodels.ElevaVINResponse, error) {
	if err := e.getAccessToken(); err != nil {
		return nil, err
	}

	url := "https://api-prd-js.eleva-services.com/v4/showcase-vehicle/client-summary"
	payload := map[string]string{
		"plateOrChassis": plateOrVIN,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating vin info request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.accessToken)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing vin info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vin info failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading vin info response: %w", err)
	}
	v := &coremodels.ElevaVINResponse{}
	err = json.Unmarshal(jsonResp, &v)
	if err != nil {
		return nil, fmt.Errorf("failed decoding vin info response: %w", err)
	}

	return v, nil
}
