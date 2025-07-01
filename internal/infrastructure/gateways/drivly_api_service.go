package gateways

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/pkg/errors"
)

//go:generate mockgen -source drivly_api_service.go -destination mocks/drivly_api_service_mock.go -package mocks
type DrivlyAPIService interface {
	GetVINInfo(vin string) (*coremodels.DrivlyVINResponse, error)
}

type drivlyAPIService struct {
	settings      *config.Settings
	httpClientVIN http.ClientWrapper
}

func NewDrivlyAPIService(settings *config.Settings) DrivlyAPIService {
	if settings.DrivlyAPIKey == "" {
		panic("Drivly configuration not set")
	}
	h := map[string]string{"x-api-key": settings.DrivlyAPIKey}
	hcwv, _ := http.NewClientWrapper(settings.DrivlyVINAPIURL.String(), "", 10*time.Second, h, true, http.WithRetry(1))

	return &drivlyAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
	}
}

// GetVINInfo is the basic enriched VIN call, that is pretty standard now. Looks in multiple sources in their backend.
func (ds *drivlyAPIService) GetVINInfo(vin string) (*coremodels.DrivlyVINResponse, error) {
	res, err := executeAPI(ds.httpClientVIN, fmt.Sprintf("/api/%s/", vin))
	if err != nil {
		return nil, err
	}
	v := &coremodels.DrivlyVINResponse{}
	err = json.Unmarshal(res, v)
	if err != nil {
		return nil, err
	}

	if v.Year == "0" || len(v.Year) == 0 || len(v.Model) == 0 || len(v.Make) == 0 {
		return nil, fmt.Errorf("decode failed due to invalid MMY. Make: %s Model: %s Year: %s", v.Make, v.Model, v.Year)
	}

	return v, nil
}

func executeAPI(httpClient http.ClientWrapper, path string) ([]byte, error) {
	res, err := httpClient.ExecuteRequest(path, "GET", nil)
	if res == nil {
		if err != nil {
			return nil, errors.Wrapf(err, "error calling driv.ly api => %s", path)
		}
		return nil, fmt.Errorf("received error with no response when calling GET to %s", path)
	}

	if err != nil && res.StatusCode != 404 {
		return nil, errors.Wrapf(err, "error calling api => %s", path)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	return body, nil
}
