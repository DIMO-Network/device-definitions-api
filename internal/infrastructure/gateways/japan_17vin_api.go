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

// bodyStylePattern matches Japanese EPC body-style codes like "4D", "5D", "2D",
// "4DR", "5HB", "4WD", which sometimes appear in the "Model Name" column instead
// of an actual vehicle series.
//
// The shape is deliberately narrow: a single door/drive count followed by one of
// a closed set of body-type abbreviations -- D/DR (door), HB (hatchback), W/WG
// (wagon), WD (wheel drive), WS (wheel steering). An earlier version was
// `^\d+[A-Za-z]{0,3}$`, which also swallowed genuine model names that happen to
// be digit-led: 86, 911, 500, 300, 240, 370Z, 350Z, 240SX, 2000GT. Those decoded
// to an empty model and then to an unresolvable definition id. Widening the
// letter count to {1,3} is not sufficient either -- it still eats 370Z, 240SX and
// 2000GT -- so the letters have to be enumerated rather than counted.
var bodyStylePattern = regexp.MustCompile(`^[1-9](D|DR|HB|W|WG|WD|WS)$`)

// nonModelTokens are EPC markers observed in the "Additional Vehicle
// Infomation" column that are never a vehicle series. They only matter on the
// last-resort path in extractModelName, which otherwise takes the column's first
// multi-character token and would happily return "LHD" as a model.
var nonModelTokens = map[string]struct{}{
	"LHD":  {}, // left-hand drive
	"RHD":  {}, // right-hand drive
	"CBU":  {}, // completely built up
	"DCB":  {}, // double cab
	"SED":  {}, // sedan
	"HB":   {}, // hatchback
	"WGN":  {}, // wagon
	"HTWC": {}, // EPC trim marker
	"TBO":  {}, // turbo
	"USA":  {}, // market
	"CHI":  {}, // market
	"UK":   {}, // market
}

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

// extractModelName finds the vehicle series name from the 17vin response.
// Preference order: (1) data.model_list[0].Model_en — 17vin's standardized
// product name, populated for known 17-char VINs; (2) the EPC
// model_original_epc_list attributes, which may carry raw internal model
// names with body-style codes, parenthetical factory tags, or multi-model
// platform lists (e.g. "NOAH/VOXY/ESQUIRE").
func extractModelName(parsed gjson.Result) string {
	if std := sanitizeModelName(parsed.Get("data.model_list.0.Model_en").String()); std != "" {
		return std
	}
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
				if isModelCandidateToken(t) {
					return sanitizeModelName(t)
				}
			}
		}
	}
	return ""
}

// sanitizeModelName strips characters that don't belong in a slugged model
// name: parenthetical factory tags like "(TMMC" and trailing commas.
func sanitizeModelName(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "(,"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// pickModelCandidate accepts a raw Col_value, splits on "/" for multi-model responses,
// and returns the best non-body-style entry. If multiple candidates remain and the
// hint (additional-info column) contains one of them, that one wins; otherwise the
// first valid candidate is returned. Parenthetical suffixes are stripped.
func pickModelCandidate(raw, hint string) string {
	raw = sanitizeModelName(raw)
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

// isModelCandidateToken reports whether a bare token from the additional-info
// column could plausibly be a vehicle series. It must be longer than one
// character, must contain a letter, and must be neither a body-style code nor a
// known EPC marker. Trim/seating codes such as "05S" and "07S" are rejected by
// the letter-position check: a series name does not lead with a digit and end in
// a single letter.
func isModelCandidateToken(t string) bool {
	t = strings.TrimSpace(t)
	if len(t) < 2 || isBodyStyleCode(t) {
		return false
	}
	if _, ok := nonModelTokens[strings.ToUpper(t)]; ok {
		return false
	}
	return strings.ContainsFunc(t, unicode.IsLetter) && !seatingCodePattern.MatchString(t)
}

// seatingCodePattern matches EPC seating/spec codes like "05S", "07S", "08S".
var seatingCodePattern = regexp.MustCompile(`^\d{2}[A-Za-z]$`)

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
