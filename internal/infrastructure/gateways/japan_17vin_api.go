package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
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

// bodyStylePattern matches Japanese EPC body-style codes like "4D", "5D", "2D", "4DR", "5HB"
// which sometimes appear in the "Model Name" column instead of an actual vehicle series.
var bodyStylePattern = regexp.MustCompile(`^\d+[A-Za-z]{0,3}$`)

// modelNameColumns lists Col_name candidates for the vehicle series, in priority order.
// 17vin responses vary by brand; docs show "Model name" (lowercase), existing production
// data used "Model Name", and Chinese-column responses use "车型".
var modelNameColumns = []string{"Model Name", "Model name", "车型"}

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
	model := extractModelName(parsed)

	result := coremodels.Japan17MMY{
		VIN:                   vin,
		ManufacturerName:      capitalize(parsed.Get("data.epc").String()),
		ManufacturerLowerCase: parsed.Get("data.epc").String(),
		ModelName:             model,
		Year:                  year,
	}

	return &result, bodyBytes, nil
}

// extractModelName finds the vehicle series name from the 17vin EPC response,
// skipping body-style codes like "4D" that sometimes populate the Model Name column.
// When Model Name is a "/"-joined platform list (e.g. "NOAH/VOXY/ESQUIRE"), the
// Additional Vehicle Infomation field usually names the actual trim — used as
// a disambiguation hint.
func extractModelName(parsed gjson.Result) string {
	entries := parsed.Get("data.model_original_epc_list").Array()
	for _, entry := range entries {
		attrs := entry.Get("CarAttributes")
		addl := attrs.Get(`#(Col_name="Additional Vehicle Infomation").Col_value`).String()
		for _, col := range modelNameColumns {
			raw := attrs.Get(fmt.Sprintf(`#(Col_name=%q).Col_value`, col)).String()
			candidate := pickModelCandidate(raw, addl)
			if candidate != "" {
				return candidate
			}
		}
		// fall back: first meaningful token of Additional Vehicle Infomation
		if addl != "" {
			for t := range strings.FieldsSeq(addl) {
				if !isBodyStyleCode(t) && len(t) > 1 {
					return t
				}
			}
		}
	}
	return ""
}

// pickModelCandidate accepts a raw Col_value, splits on "/" for multi-model responses,
// and returns the best non-body-style entry. If multiple candidates remain and the
// hint (additional-info column) contains one of them, that one wins; otherwise the
// first valid candidate is returned.
func pickModelCandidate(raw, hint string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var candidates []string
	for p := range strings.SplitSeq(raw, "/") {
		p = strings.TrimSpace(p)
		if p != "" && !isBodyStyleCode(p) {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) > 1 && hint != "" {
		hintUpper := strings.ToUpper(hint)
		for _, c := range candidates {
			if strings.Contains(hintUpper, strings.ToUpper(c)) {
				return c
			}
		}
	}
	return candidates[0]
}

func isBodyStyleCode(s string) bool {
	return bodyStylePattern.MatchString(strings.TrimSpace(s))
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
