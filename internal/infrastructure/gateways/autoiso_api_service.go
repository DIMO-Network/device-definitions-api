package gateways

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"

	"github.com/DIMO-Network/shared/pkg/http"
)

//go:generate mockgen -source autoiso_api_service.go -destination mocks/autoiso_api_service_mock.go -package mocks
type AutoIsoAPIService interface {
	GetVIN(vin string) (*coremodels.AutoIsoVINResponse, []byte, error)
}

type autoIsoAPIService struct {
	httpClientVIN http.ClientWrapper
	autoIsoAPIUid string
	autoIsoAPIKey string
}

func NewAutoIsoAPIService(autoIsoAPIUid, autoIsoAPIKey string) AutoIsoAPIService {
	if autoIsoAPIUid == "" || autoIsoAPIKey == "" {
		panic("Drivly configuration not set")
	}
	hcwv, _ := http.NewClientWrapper("http://bp.autoiso.pl", "", 10*time.Second, nil, false)

	return &autoIsoAPIService{
		httpClientVIN: hcwv,
		autoIsoAPIUid: autoIsoAPIUid,
		autoIsoAPIKey: autoIsoAPIKey,
	}
}

func (ai *autoIsoAPIService) GetVIN(vin string) (*coremodels.AutoIsoVINResponse, []byte, error) {
	input := ai.autoIsoAPIUid + ai.autoIsoAPIKey + vin
	// has with md5
	hasher := md5.New()
	hasher.Write([]byte(input))
	hashedBytes := hasher.Sum(nil)
	hashedChecksum := hex.EncodeToString(hashedBytes)

	res, err := executeAPI(ai.httpClientVIN, fmt.Sprintf("/api/v3/getSimpleDecoder/apiuid:DIMOZ/checksum:%s/vin:%s", hashedChecksum, vin))
	if err != nil {
		return nil, nil, err
	}
	v := &coremodels.AutoIsoVINResponse{}
	err = json.Unmarshal(res, v)
	if err != nil {
		return nil, res, err
	}
	// get percent match from autoiso, if below 50 return err - kinda of an experiment for now
	percentMatchStr := strings.TrimSuffix(v.FunctionResponse.Data.API.DataMatching, "%")
	percentMatch, _ := strconv.ParseFloat(percentMatchStr, 64)
	if percentMatch < 55.0 {
		return nil, res, fmt.Errorf("decode failed due to low DataMatching percent: %f. MMY: %s %s %s", percentMatch,
			v.FunctionResponse.Data.Decoder.Make.Value, v.FunctionResponse.Data.Decoder.Model.Value, v.FunctionResponse.Data.Decoder.ModelYear.Value)
	}

	if v.FunctionResponse.Data.Decoder.ModelYear.Value == "0" || len(v.FunctionResponse.Data.Decoder.ModelYear.Value) == 0 ||
		len(v.FunctionResponse.Data.Decoder.Model.Value) == 0 || len(v.FunctionResponse.Data.Decoder.Make.Value) == 0 {
		return nil, res, fmt.Errorf("decode failed due to invalid MMY. Make: %s Model: %s Year: %s", v.FunctionResponse.Data.Decoder.Make.Value,
			v.FunctionResponse.Data.Decoder.Model.Value, v.FunctionResponse.Data.Decoder.ModelYear.Value)
	}

	return v, res, nil
}
