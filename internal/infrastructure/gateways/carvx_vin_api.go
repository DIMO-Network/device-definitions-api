package gateways

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source carvx_vin_api.go -destination mocks/carvx_vin_api_mock.go -package mocks
type carVxVINAPI struct {
	httpClient http.ClientWrapper
	logger     zerolog.Logger
	settings   *config.Settings
}

type CarVxVINAPI interface {
	GetVINInfo(chassisNumber string) (*coremodels.CarVxResponse, []byte, error)
}

const carvxURL = "https://carvx.jp/api/v1/get-chassis-info"

func NewCarVxVINAPI(logger zerolog.Logger, settings *config.Settings) CarVxVINAPI {
	headers := map[string]string{
		"Carvx-User-Uid": settings.CarVxUserID,
		"Carvx-Api-Key":  settings.CarVxAPIKey,
	}
	hc, _ := http.NewClientWrapper(carvxURL, "", 15*time.Second, headers, true, http.WithRetry(2))
	return &carVxVINAPI{
		httpClient: hc,
		logger:     logger,
		settings:   settings,
	}
}

func (c *carVxVINAPI) GetVINInfo(chassisNumber string) (*coremodels.CarVxResponse, []byte, error) {
	response, err := c.httpClient.ExecuteRequest(fmt.Sprintf("?chassis_number=%s", chassisNumber), "GET", nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to decode chassis number info from carvx api: %s", chassisNumber)
	}
	defer response.Body.Close() //nolint
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error reading response body from url %s", carvxURL)
	}
	v := &coremodels.CarVxResponse{}
	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error decoding response body from url %s", carvxURL)
	}
	if len(v.Error) > 0 {
		return nil, nil, errors.New(v.Error)
	}
	if len(v.Data) == 0 {
		return nil, nil, errors.New("no data found")
	}
	return v, bodyBytes, nil
}
