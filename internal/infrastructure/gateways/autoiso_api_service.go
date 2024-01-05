package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
)

//go:generate mockgen -source autoiso_api_service.go -destination mocks/autoiso_api_service_mock.go -package mocks
type AutoIsoAPIService interface {
	GetVIN(vin string) (*AutoIsoVINResponse, error)
}

type autoIsoAPIService struct {
	Settings      *config.Settings
	httpClientVIN shared.HTTPClientWrapper
}

func NewAutoIsoAPIService(settings *config.Settings) AutoIsoAPIService {
	if settings.AutoIsoAPIUid == "" || settings.AutoIsoAPIKey == "" {
		panic("Drivly configuration not set")
	}
	hcwv, _ := shared.NewHTTPClientWrapper("http://bp.autoiso.pl", "", 10*time.Second, nil, false)

	return &autoIsoAPIService{
		Settings:      settings,
		httpClientVIN: hcwv,
	}
}

func (ai *autoIsoAPIService) GetVIN(vin string) (*AutoIsoVINResponse, error) {
	input := ai.Settings.AutoIsoAPIUid + ai.Settings.AutoIsoAPIKey + vin
	// has with md5
	hasher := md5.New()
	hasher.Write([]byte(input))
	hashedBytes := hasher.Sum(nil)
	hashedChecksum := hex.EncodeToString(hashedBytes)

	res, err := executeAPI(ai.httpClientVIN, fmt.Sprintf("/api/v3/getSimpleDecoder/apiuid:DIMOZ/checksum:%s/vin:%s", hashedChecksum, vin))
	if err != nil {
		return nil, err
	}
	v := &AutoIsoVINResponse{}
	err = json.Unmarshal(res, v)
	if err != nil {
		return nil, err
	}
	// get percent match from autoiso, if below 50 return err - kinda of an experiment for now
	percentMatchStr := strings.TrimSuffix(v.FunctionResponse.Data.API.DataMatching, "%")
	percentMatch, err := strconv.ParseFloat(percentMatchStr, 64)
	if percentMatch < 50.0 {
		return nil, fmt.Errorf("decode failed due to low DataMatching percent: %f", percentMatch)
	}

	if v.FunctionResponse.Data.Decoder.ModelYear.Value == "0" || len(v.FunctionResponse.Data.Decoder.ModelYear.Value) == 0 ||
		len(v.FunctionResponse.Data.Decoder.Model.Value) == 0 || len(v.FunctionResponse.Data.Decoder.Make.Value) == 0 {
		return nil, fmt.Errorf("decode failed due to invalid MMY")
	}

	return v, nil
}

type AutoIsoVINResponse struct {
	Version          string `json:"version"`
	Vin              string `json:"vin"`
	APIStatus        string `json:"apiStatus"`
	ResponseDate     string `json:"responseDate"`
	FunctionName     string `json:"functionName"`
	FunctionResponse struct {
		Data struct {
			API struct {
				CoreVersion     string `json:"core_version"`
				EndpointVersion int    `json:"endpoint_version"`
				JSONVersion     string `json:"json_version"`
				APIType         string `json:"api_type"`
				APICache        string `json:"api_cache"`
				DataPrecision   int    `json:"data_precision"`
				DataMatching    string `json:"data_matching"`
				LexLang         string `json:"lex_lang"`
			} `json:"api"`
			Analyze struct {
				VinOrginal struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_orginal"`
				VinCorrected struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_corrected"`
				VinYear struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_year"`
				VinSerial struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vin_serial"`
				Checkdigit struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"checkdigit"`
			} `json:"analyze"`
			Decoder struct {
				Make struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"make"`
				Model struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"model"`
				ModelYear struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"model_year"`
				Body struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"body"`
				FuelType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"fuel_type"`
				VehicleType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"vehicle_type"`
				Doors struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"doors"`
				EngineDisplCm3 struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_displ_cm3"`
				EngineDisplL struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_displ_l"`
				EnginePowerHp struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_power_hp"`
				EnginePowerKw struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_power_kw"`
				EngineConf struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_conf"`
				EngineType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_type"`
				EngineVersion struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_version"`
				EngineHead struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_head"`
				EngineValves struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_valves"`
				EngineCylinders struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_cylinders"`
				EngineDisplCid struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_displ_cid"`
				EngineTurbo struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"engine_turbo"`
				DriveType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"drive_type"`
				GearboxType struct {
					Desc  string `json:"desc"`
					Value string `json:"value"`
				} `json:"gearbox_type"`
				EmissionStd struct {
					Desc   string `json:"desc"`
					Value  string `json:"value"`
					Co2Gkm string `json:"co2_gkm"`
				} `json:"emission_std"`
			} `json:"decoder"`
		} `json:"data"`
	} `json:"functionResponse"`
	LicenseInfo struct {
		LicenseNumber         string `json:"licenseNumber"`
		ValidTo               string `json:"validTo"`
		RemainingCredits      int    `json:"remainingCredits"`
		RemainingMonthlyLimit int    `json:"remainingMonthlyLimit"`
		RemainingDailyLimit   int    `json:"remainingDailyLimit"`
	} `json:"licenseInfo"`
}

// GetStyle returns a standard style string built from the data we have
func (v *AutoIsoVINResponse) GetStyle() string {
	s := ""
	// eg. Diesel
	if len(v.FunctionResponse.Data.Decoder.FuelType.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.FuelType.Value + " "
	}
	// eg. TDI
	if len(v.FunctionResponse.Data.Decoder.EngineVersion.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.EngineVersion.Value + " "
	}
	// eg. Manual
	if len(v.FunctionResponse.Data.Decoder.GearboxType.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.GearboxType.Value + " "
	}
	// eg. FWD / AWD
	if len(v.FunctionResponse.Data.Decoder.DriveType.Value) > 0 {
		s += v.FunctionResponse.Data.Decoder.DriveType.Value + " "
	}

	return strings.TrimSpace(s)
}

// GetSubModel returns the Body type, which we can use as the sub model.
func (v *AutoIsoVINResponse) GetSubModel() string {
	return strings.TrimSpace(v.FunctionResponse.Data.Decoder.Body.Value)
}
