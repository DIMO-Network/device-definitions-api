package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"time"
)

type Japan17VINAPI interface {
	GetVINInfo(vin string) (*coremodels.Japan17VINResp, error)
}

type japan17VINAPI struct {
	logger     *zerolog.Logger
	settings   *config.Settings
	httpClient shared.HTTPClientWrapper
}

func NewJapan17VINAPI(logger *zerolog.Logger, settings *config.Settings) Japan17VINAPI {
	httpClient, _ := shared.NewHTTPClientWrapper("", "", 20*time.Second, nil, true)

	return &japan17VINAPI{
		logger:     logger,
		settings:   settings,
		httpClient: httpClient,
	}
}

func (j *japan17VINAPI) GetVINInfo(vin string) (*coremodels.Japan17VINResp, error) {
	token := tokenGenerator(j.settings.Japan17VINUser, j.settings.Japan17VINPassword, vin)

	url := fmt.Sprintf("http://api.17vin.com:8080/?vin=%s&user=%s&token=%s", vin, j.settings.Japan17VINUser, token)

	response, err := j.httpClient.ExecuteRequest(url, "GET", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get VIN info from 17vin api: %s", vin)
	}
	defer response.Body.Close() //nolint

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading response body from url %s", url)
	}
	var result coremodels.Japan17VINResp

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
func md5Hex(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func tokenGenerator(user, password, vin string) string {
	// from https://www.17vin.com/doc.aspx

	usernameHash := md5Hex(user)
	passwordHash := md5Hex(password)

	// Step 2: Concatenate hashes with "/?vin=..."
	combined := usernameHash + passwordHash + "/?vin=" + vin

	// Step 3: MD5 of the whole string
	finalHash := md5Hex(combined)
	return finalHash
}
