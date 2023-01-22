package elasticsearch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/rs/zerolog"
)

type ElasticSearch struct {
	BaseURL    string
	httpClient shared.HTTPClientWrapper
	log        zerolog.Logger
}

func NewElasticSearch(settings *config.Settings, logger zerolog.Logger) (*ElasticSearch, error) {
	headers := map[string]string{
		"Authorization": "ApiKey " + settings.ElasticSearchDeviceStatusToken,
	}
	client, err := shared.NewHTTPClientWrapper("", "", 240*time.Second, headers, true)
	if err != nil {
		return nil, err
	}

	return &ElasticSearch{
		BaseURL:    settings.ElasticSearchDeviceStatusHost,
		httpClient: client,
		log:        logger,
	}, nil
}

// buildAndExecRequest makes request with token and headers, marshalling passed in obj or if string just passing along in body,
// and unmarshalling response body to objOut (must be passed in as ptr eg &varName)
func (d *ElasticSearch) buildAndExecRequest(method, url string, obj interface{}, objOut interface{}) error {
	var reqBody []byte

	if obj != nil {
		if s, ok := obj.(string); ok {
			reqBody = []byte(s)
		} else {
			var err error
			reqBody, err = json.Marshal(obj)
			if err != nil {
				return fmt.Errorf("failed marshaling request body: %w", err)
			}
		}
	}

	resp, err := d.httpClient.ExecuteRequest(url, method, reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if objOut != nil {
		err = json.NewDecoder(resp.Body).Decode(objOut)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
