package gateways

import (
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"time"
)

type carVxVINAPI struct {
	httpClient http.ClientWrapper
	logger     zerolog.Logger
	settings   *config.Settings
}

type CarVxVINAPI interface {
	GetVINInfo(chassisNumber string) (*coremodels.CarVxResponse, error)
}

const carvxURL = "https://carvx.jp/api/v1/get-chassis-info"

func NewCarVxVINAPI(logger zerolog.Logger, settings *config.Settings) CarVxVINAPI {
	headers := map[string]string{
		"Carvx-User-Uid": settings.CarVxUserId,
		"Carvx-Api-Key":  settings.CarVxAPIKey,
	}
	hc, _ := http.NewClientWrapper(carvxURL, "", 15*time.Second, headers, true, http.WithRetry(2))
	return &carVxVINAPI{
		httpClient: hc,
		logger:     logger,
		settings:   settings,
	}
}

func (c *carVxVINAPI) GetVINInfo(chassisNumber string) (*coremodels.CarVxResponse, error) {
	response, err := c.httpClient.ExecuteRequest(fmt.Sprintf("?chassis_number=%s", chassisNumber), "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode chassis number info from carvx api: %s", chassisNumber)
	}
	defer response.Body.Close() //nolint
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading response body from url %s", carvxURL)
	}
	v := &coremodels.CarVxResponse{}
	err = json.Unmarshal(bodyBytes, v)
	if err != nil {
		return nil, errors.Wrapf(err, "error decoding response body from url %s", carvxURL)
	}
	return v, nil
}
