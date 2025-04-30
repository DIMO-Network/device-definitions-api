//nolint:tagliatelle
package gateways

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"reflect"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

//go:generate mockgen -source vincario_api_service.go -destination mocks/vincario_api_service_mock.go -package mocks
type VincarioAPIService interface {
	DecodeVIN(vin string) (*coremodels.VincarioInfoResponse, error)
}

type vincarioAPIService struct {
	settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
	log           *zerolog.Logger
}

func NewVincarioAPIService(settings *config.Settings, log *zerolog.Logger) VincarioAPIService {
	if settings.VincarioAPISecret == "" {
		panic("Vincario configuration not set")
	}
	hcwv, _ := shared.NewHTTPClientWrapper(settings.VincarioAPIURL.String(), "", 10*time.Second, nil, false)

	return &vincarioAPIService{
		settings:      settings,
		httpClientVIN: hcwv,
		log:           log,
	}
}

func (va *vincarioAPIService) DecodeVIN(vin string) (*coremodels.VincarioInfoResponse, error) {
	id := "decode"

	urlPath := vincarioPathBuilder(vin, id, va.settings.VincarioAPIKey, va.settings.VincarioAPISecret)
	// url with api access
	resp, err := va.httpClientVIN.ExecuteRequest(urlPath, "GET", nil)
	if err != nil {
		return nil, err
	}

	// decode JSON from response body
	var data tempResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	infoResp := coremodels.VincarioInfoResponse{}
	for _, s2 := range data.Decode {
		// find the property in the struct with the label name in the key metadata
		err := setStructPropertiesByMetadataKey(&infoResp, s2.Label, s2.Value)
		if err != nil {
			va.log.Debug().Err(err).Msg("could not set struct properties")
		}
	}

	if infoResp.ModelYear == 0 || len(infoResp.Model) == 0 || len(infoResp.Make) == 0 {
		return nil, fmt.Errorf("decode failed due to invalid MMY. Make %s, Model %s, Year %d", infoResp.Make, infoResp.Model, infoResp.ModelYear)
	}

	return &infoResp, nil
}

func vincarioPathBuilder(vin, id, key, secret string) string {
	s := vin + "|" + id + "|" + key + "|" + secret

	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)

	controlSum := hex.EncodeToString(bs[0:5])

	return "/" + key + "/" + controlSum + "/" + id + "/" + vin + ".json"
}

func setStructPropertiesByMetadataKey(structPtr interface{}, key string, value interface{}) error {
	structValue := reflect.ValueOf(structPtr).Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		if fieldType.Tag.Get("key") == key {
			if !field.CanSet() {
				return fmt.Errorf("field %s is unexported and cannot be set", fieldType.Name)
			}

			fieldValue := reflect.ValueOf(value)

			if !fieldValue.Type().AssignableTo(field.Type()) {
				if fieldValue.Kind() == reflect.Float64 || fieldValue.Kind() == reflect.Float32 {
					f := fieldValue.Float()
					if field.Kind() == reflect.Int {
						field.Set(reflect.ValueOf(int(f)))
					}
				} else {
					return fmt.Errorf("value %v is not assignable to field %s of type %s", value, fieldType.Name, field.Type())
				}
			} else {
				field.Set(fieldValue)
			}
			return nil
		}
	}

	return fmt.Errorf("struct does not contain a field with key %s", key)
}

type tempResponse struct {
	Balance struct {
		APIDecode int `json:"API Decode"`
	} `json:"balance"`
	Decode []struct {
		Label string `json:"label"`
		Value any    `json:"value"`
		ID    int    `json:"id,omitempty"`
	} `json:"decode"`
}
