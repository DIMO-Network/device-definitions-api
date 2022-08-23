package gateways

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

type ElasticSearchService interface {
	LoadDeviceDefinitions() error
	GetEngines() (*GetEnginesResp, error)
	CreateEngine(name string, metaSource *string) (*EngineDetail, error)
	AddSourceEngineToMetaEngine(sourceName, metaName string) (*EngineDetail, error)
	RemoveSourceEngine(sourceName, metaName string) (*EngineDetail, error)
	DeleteEngine(name string) error
	CreateDocuments(docs []DeviceDefinitionSearchDoc, engineName string) ([]CreateDocsResp, error)
	CreateDocumentsBatched(docs []DeviceDefinitionSearchDoc, engineName string) error
	UpdateSearchSettingsForDeviceDefs(engineName string) error
}

type elasticSearchService struct {
	BaseURL        string
	Token          string
	MetaEngineName string
	httpClient     *http.Client
	log            zerolog.Logger
}

func NewElasticSearchService(settings *config.Settings, logger zerolog.Logger) (ElasticSearchService, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	return &elasticSearchService{
		BaseURL:        settings.ElasticSearchAppSearchHost,
		Token:          settings.ElasticSearchAppSearchToken,
		MetaEngineName: "dd-" + settings.Environment,
		httpClient:     netClient,
		log:            logger,
	}, nil
}

func (d *elasticSearchService) LoadDeviceDefinitions() error {
	return nil
}

// GetEngines Calls elastic instance api to list engines
func (d *elasticSearchService) GetEngines() (*GetEnginesResp, error) {
	url := fmt.Sprintf("%s/api/as/v1/engines/", d.BaseURL)
	engines := GetEnginesResp{}
	_, err := d.buildAndExecRequest("GET", url, nil, &engines)
	if err != nil {
		return nil, errors.Wrap(err, "error getting engines")
	}

	return &engines, nil
}

// CreateEngine https://www.elastic.co/guide/en/app-search/current/engines.html#engines-create
func (d *elasticSearchService) CreateEngine(name string, metaSource *string) (*EngineDetail, error) {
	if strings.ToLower(name) != name {
		return nil, errors.New("name must be all lowercase")
	}

	url := fmt.Sprintf("%s/api/as/v1/engines/", d.BaseURL)
	lang := "Universal"
	meta := "meta"
	create := EngineDetail{
		Name:     name,
		Language: &lang,
	}
	if metaSource != nil {
		create.Language = nil
		create.Type = &meta
		create.SourceEngines = []string{*metaSource}
	}
	engine := EngineDetail{}
	_, err := d.buildAndExecRequest("POST", url, create, &engine)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating engine: %v", create)
	}

	return &engine, nil
}

// AddSourceEngineToMetaEngine https://www.elastic.co/guide/en/app-search/current/meta-engines.html#meta-engines-add-source-engines
func (d *elasticSearchService) AddSourceEngineToMetaEngine(sourceName, metaName string) (*EngineDetail, error) {
	url := fmt.Sprintf("%s/api/as/v1/engines/%s/source_engines", d.BaseURL, metaName)
	body := `["%s"]`
	body = fmt.Sprintf(body, sourceName)

	engine := EngineDetail{}
	_, err := d.buildAndExecRequest("POST", url, body, &engine)
	if err != nil {
		return nil, errors.Wrapf(err, "error adding source engine: %s to meta engine: %s", sourceName, metaName)
	}

	return &engine, nil
}

// RemoveSourceEngine https://www.elastic.co/guide/en/app-search/current/meta-engines.html#meta-engines-remove-source-engines
func (d *elasticSearchService) RemoveSourceEngine(sourceName, metaName string) (*EngineDetail, error) {
	url := fmt.Sprintf("%s/api/as/v1/engines/%s/source_engines", d.BaseURL, metaName)
	body := `["%s"]`
	body = fmt.Sprintf(body, sourceName)

	engine := EngineDetail{}
	_, err := d.buildAndExecRequest("DELETE", url, body, &engine)
	if err != nil {
		return nil, errors.Wrapf(err, "error removing source engine: %s from meta engine: %s", sourceName, metaName)
	}

	return &engine, nil
}

// DeleteEngine https://www.elastic.co/guide/en/app-search/current/engines.html#engines-delete
func (d *elasticSearchService) DeleteEngine(name string) error {
	url := fmt.Sprintf("%s/api/as/v1/engines/%s", d.BaseURL, name)
	// DELETE
	_, err := d.buildAndExecRequest("DELETE", url, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting engine %s", name)
	}
	return nil
}

// CreateDocuments https://www.elastic.co/guide/en/app-search/current/documents.html#documents-create
// max of 100 documents per batch, 100kb each document
func (d *elasticSearchService) CreateDocuments(docs []DeviceDefinitionSearchDoc, engineName string) ([]CreateDocsResp, error) {
	// todo: make docs generic parameter?
	if len(docs) > 100 {
		return nil, fmt.Errorf("docs slice is too big with len: %d, max of 100 items allowed", len(docs))
	}
	url := fmt.Sprintf("%s/api/as/v1/engines/%s/documents", d.BaseURL, engineName)

	var resp []CreateDocsResp
	_, err := d.buildAndExecRequest("POST", url, docs, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating documents in engine: %s", engineName)
	}
	// todo: what about iterating over the resp errors property to log that?
	return resp, nil
}

// CreateDocumentsBatched splits chunks of 100 docs and calls CreateDocuments method with each chunk
func (d *elasticSearchService) CreateDocumentsBatched(docs []DeviceDefinitionSearchDoc, engineName string) error {
	chunkSize := 100
	prev := 0
	i := 0
	till := len(docs) - chunkSize
	for prev < till {
		next := prev + chunkSize
		_, err := d.CreateDocuments(docs[prev:next], engineName)
		if err != nil {
			return err
		}
		prev = next
		i++
	}
	// remainder
	_, err := d.CreateDocuments(docs[prev:], engineName)
	return err
}

// UpdateSearchSettingsForDeviceDefs specific method to update the search_settings for device definitions
// https://www.elastic.co/guide/en/app-search/current/search-settings.html#search-settings-update
func (d *elasticSearchService) UpdateSearchSettingsForDeviceDefs(engineName string) error {
	url := fmt.Sprintf("%s/api/as/v1/engines/%s/search_settings", d.BaseURL, engineName)
	body := `
{
  "search_fields": {
    "search_display": {
      "weight": 1
    },
    "sub_models": {
      "weight": 0.7
    }
  },
  "result_fields": {
    "year": {
      "raw": {}
    },
    "image_url": {
      "raw": {}
    },
    "search_display": {
      "raw": {}
    },
    "id": {
      "raw": {}
    },
    "model": {
      "raw": {}
    },
    "sub_models": {
      "raw": {}
    },
    "make": {
      "raw": {}
    }
  },
  "boosts": {},
  "precision": 2
}`
	_, err := d.buildAndExecRequest("PUT", url, body, nil)
	if err != nil {
		return errors.Wrapf(err, "error when trying to update search_settings for: %s", engineName)
	}
	return nil
}

// buildAndExecRequest makes request with token and headers, marshalling passed in obj or if string just passing along in body,
// and unmarshalling response body to objOut (must be passed in as ptr eg &varName)
func (d *elasticSearchService) buildAndExecRequest(method, url string, obj interface{}, objOut interface{}) (*http.Response, error) {
	backoffSchedule := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		10 * time.Second,
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
	req.Header.Set("Authorization", "Bearer "+d.Token)
	var resp *http.Response
	var err error

	for _, backoff := range backoffSchedule {
		resp, err = d.httpClient.Do(req) // any error resp should return err per docs
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
		err = json.NewDecoder(resp.Body).Decode(&objOut)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding response json")
		}
	}

	return resp, nil
}

// DeviceDefinitionSearchDoc used for elastic search document indexing. entirely for searching, source of truth is DB.
// elastic only supports lowercase letters, number and underscores, ie. snake_case
type DeviceDefinitionSearchDoc struct {
	ID string `json:"id"`
	// SearchDisplay M+M+Y
	SearchDisplay string `json:"search_display"`
	Make          string `json:"make"`
	Model         string `json:"model"`
	Year          int    `json:"year"`
	// SubModels: M+M+Y+Submodel name
	SubModels []string `json:"sub_models"`
	ImageURL  string   `json:"image_url"`
	// future: we might add styles
}

type GetEnginesResp struct {
	Meta struct {
		Page struct {
			Current      int `json:"current"`
			TotalPages   int `json:"total_pages"`
			TotalResults int `json:"total_results"`
			Size         int `json:"size"`
		} `json:"page"`
	} `json:"meta"`
	Results []EngineDetail `json:"results"`
}

// EngineDetail can be used as a response when listing engines, or to create an engine. The minimum parameters are Name. Type and Source
// can be used when dealing with Meta Engines: https://www.elastic.co/guide/en/app-search/current/meta-engines.html
type EngineDetail struct {
	Name          string   `json:"name"`
	Language      *string  `json:"language,omitempty"`
	Type          *string  `json:"type,omitempty"`
	DocumentCount *int     `json:"document_count,omitempty"`
	SourceEngines []string `json:"source_engines,omitempty"`
}

type CreateDocsResp struct {
	ID     string   `json:"id"`
	Errors []string `json:"errors"`
}
