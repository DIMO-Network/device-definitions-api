package elasticsearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type ElasticSearch struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
	log        zerolog.Logger
}

func NewElasticSearch(settings *config.Settings, logger zerolog.Logger) (*ElasticSearch, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 20,
	}
	return &ElasticSearch{
		BaseURL:    settings.ElasticSearchDeviceStatusHost,
		Token:      settings.ElasticSearchDeviceStatusToken,
		httpClient: netClient,
		log:        logger,
	}, nil
}

// buildAndExecRequest makes request with token and headers, marshalling passed in obj or if string just passing along in body,
// and unmarshalling response body to objOut (must be passed in as ptr eg &varName)
func (d *ElasticSearch) buildAndExecRequest(method, url string, obj interface{}, objOut interface{}) (*http.Response, error) {
	backoffSchedule := []time.Duration{
		3 * time.Second,
		10 * time.Second,
		30 * time.Second,
	}

	var req *http.Request

	if obj == nil {
		req, _ = http.NewRequest(
			method,
			url,
			nil,
		)
	} else {
		b := ""
		if reflect.TypeOf(obj).Name() == "string" {
			b = obj.(string)
		} else {
			j, _ := json.Marshal(obj)
			b = string(j)
		}
		req, _ = http.NewRequest(
			method,
			url,
			strings.NewReader(b),
		)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "ApiKey "+d.Token)
	var resp *http.Response
	var err error
	var nilRespErr error

	for _, backoff := range backoffSchedule {
		resp, nilRespErr = d.httpClient.Do(req) // any error resp should return err per docs
		if resp != nil && resp.StatusCode == http.StatusOK && err == nil {
			break
		}
		if resp != nil && resp.StatusCode == http.StatusBadRequest {
			b, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return resp, fmt.Errorf("received bad request response with body: %s", string(b))
		}
		// control for err or resp being nil to log message.
		respStatus := ""
		errMsg := ""
		if resp != nil {
			respStatus = resp.Status
		}
		if err != nil {
			if resp != nil {
				b, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if err == nil {
					errMsg = string(b)
				}
			} else {
				errMsg = err.Error()
			}
		}
		d.log.Warn().Msgf("request Status: %s. error: %s. Retrying in %v", respStatus, errMsg, backoff)
		time.Sleep(backoff)
	}
	if objOut != nil {
		if resp != nil {
			err = json.NewDecoder(resp.Body).Decode(&objOut)
			if err != nil {
				return nil, errors.Wrap(err, "error decoding response json")
			}
		} else {
			return resp, nilRespErr
		}
	}

	return resp, nil
}
