package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

//go:generate mockgen -source japan_17vin_api.go -destination mocks/japan_17vin_api_mock.go -package mocks
type Japan17VINAPI interface {
	GetVINInfo(vin string) (*coremodels.Japan17MMY, []byte, error)
}

type japan17VINAPI struct {
	logger     *zerolog.Logger
	settings   *config.Settings
	httpClient http.ClientWrapper
}

func NewJapan17VINAPI(logger *zerolog.Logger, settings *config.Settings) Japan17VINAPI {
	httpClient, _ := http.NewClientWrapper("", "", 20*time.Second, nil, true, http.WithRetry(2))

	return &japan17VINAPI{
		logger:     logger,
		settings:   settings,
		httpClient: httpClient,
	}
}

func (j *japan17VINAPI) GetVINInfo(vin string) (*coremodels.Japan17MMY, []byte, error) {
	token := tokenGenerator(j.settings.Japan17VINUser, j.settings.Japan17VINPassword, vin)

	url := fmt.Sprintf("http://api.17vin.com:8080/?vin=%s&user=%s&token=%s", vin, j.settings.Japan17VINUser, token)

	response, err := j.httpClient.ExecuteRequest(url, "GET", nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get VIN info from 17vin api: %s", vin)
	}
	defer response.Body.Close() //nolint

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error reading response body from url %s", url)
	}
	parsed := gjson.ParseBytes(bodyBytes)
	yearString := parsed.Get("data.model_year_from_vin").String()
	year, err := strconv.Atoi(yearString)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse year string %s", yearString)
	}
	model := ""
	modelNameProperty := parsed.Get(`data.model_original_epc_list.0.CarAttributes.#(Col_name="Model Name").Col_value`).String()
	if modelNameProperty != "" {
		if strings.Contains(modelNameProperty, `/`) {
			// this means it has two options, sometimes in the additional info it will have the model
			additionalInfo := parsed.Get(`data.model_original_epc_list.0.CarAttributes.#(Col_name="Additional Vehicle Infomation").Col_value`).String()
			splitInfo := strings.Split(additionalInfo, " ")
			if len(splitInfo) > 0 {
				// grab the first name which is usually the model
				model = splitInfo[0]
			} else {
				// if nothing found then just grab the first value from the model name
				model = strings.Split(modelNameProperty, "/")[0]
			}
		} else {
			model = modelNameProperty
		}
	}

	result := coremodels.Japan17MMY{
		VIN:                   vin,
		ManufacturerName:      capitalize(parsed.Get("data.epc").String()),
		ManufacturerLowerCase: parsed.Get("data.epc").String(),
		ModelName:             model,
		Year:                  year,
	}

	return &result, bodyBytes, nil
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

func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
