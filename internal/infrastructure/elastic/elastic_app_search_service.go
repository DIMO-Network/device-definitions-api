//go:generate mockgen -source elastic_app_search_service.go -destination mocks/elastic_app_search_service_mock.go -package mocks

package elastic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type SearchService interface {
	LoadDeviceDefinitions() error
	GetEngines() (*GetEnginesResp, error)
	CreateEngine(name string, metaSource *string) (*EngineDetail, error)
	AddSourceEngineToMetaEngine(sourceName, metaName string) (*EngineDetail, error)
	RemoveSourceEngine(sourceName, metaName string) (*EngineDetail, error)
	DeleteEngine(name string) error
	CreateDocuments(docs []DeviceDefinitionSearchDoc, engineName string) ([]CreateDocsResp, error)
	CreateDocumentsBatched(docs []DeviceDefinitionSearchDoc, engineName string) error
	UpdateSearchSettingsForDeviceDefs(engineName string) error
	GetMetaEngineName() string
}

// This is different than regular elastic, https://www.elastic.co/guide/en/app-search/current/api-reference.html
type elasticAppSearchService struct {
	MetaEngineName string
	httpClient     shared.HTTPClientWrapper
	log            zerolog.Logger
}

func NewElasticAppSearchService(settings *config.Settings, logger zerolog.Logger) (SearchService, error) {
	headers := map[string]string{"Authorization": "Bearer " + settings.ElasticSearchAppSearchToken}
	httpClient, err := shared.NewHTTPClientWrapper(settings.ElasticSearchAppSearchHost, "", 30*time.Second, headers, true)
	if err != nil {
		return nil, err
	}

	return &elasticAppSearchService{
		MetaEngineName: "dd-" + settings.Environment,
		httpClient:     httpClient,
		log:            logger,
	}, nil
}

// GetMetaEngineName Use for testing
func (d *elasticAppSearchService) GetMetaEngineName() string {
	return d.MetaEngineName
}

func (d *elasticAppSearchService) LoadDeviceDefinitions() error {
	return nil
}

// GetEngines Calls elastic instance api to list engines
func (d *elasticAppSearchService) GetEngines() (*GetEnginesResp, error) {
	path := "/api/as/v1/engines/"
	engines := GetEnginesResp{}
	err := d.buildAndExecRequest("GET", path, nil, &engines)
	if err != nil {
		return nil, errors.Wrap(err, "error getting engines")
	}

	return &engines, nil
}

// CreateEngine https://www.elastic.co/guide/en/app-search/current/engines.html#engines-create
func (d *elasticAppSearchService) CreateEngine(name string, metaSource *string) (*EngineDetail, error) {
	if strings.ToLower(name) != name {
		return nil, errors.New("name must be all lowercase")
	}

	path := "/api/as/v1/engines/"
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
	err := d.buildAndExecRequest("POST", path, create, &engine)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating engine: %v", create)
	}

	return &engine, nil
}

// AddSourceEngineToMetaEngine https://www.elastic.co/guide/en/app-search/current/meta-engines.html#meta-engines-add-source-engines
func (d *elasticAppSearchService) AddSourceEngineToMetaEngine(sourceName, metaName string) (*EngineDetail, error) {
	path := fmt.Sprintf("/api/as/v1/engines/%s/source_engines", metaName)
	body := `["%s"]`
	body = fmt.Sprintf(body, sourceName)

	engine := EngineDetail{}
	err := d.buildAndExecRequest("POST", path, body, &engine)
	if err != nil {
		return nil, errors.Wrapf(err, "error adding source engine: %s to meta engine: %s", sourceName, metaName)
	}

	return &engine, nil
}

// RemoveSourceEngine https://www.elastic.co/guide/en/app-search/current/meta-engines.html#meta-engines-remove-source-engines
func (d *elasticAppSearchService) RemoveSourceEngine(sourceName, metaName string) (*EngineDetail, error) {
	path := fmt.Sprintf("/api/as/v1/engines/%s/source_engines", metaName)
	body := `["%s"]`
	body = fmt.Sprintf(body, sourceName)

	engine := EngineDetail{}
	err := d.buildAndExecRequest("DELETE", path, body, &engine)
	if err != nil {
		return nil, errors.Wrapf(err, "error removing source engine: %s from meta engine: %s", sourceName, metaName)
	}

	return &engine, nil
}

// DeleteEngine https://www.elastic.co/guide/en/app-search/current/engines.html#engines-delete
func (d *elasticAppSearchService) DeleteEngine(name string) error {
	path := fmt.Sprintf("/api/as/v1/engines/%s", name)
	// DELETE
	err := d.buildAndExecRequest("DELETE", path, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "error deleting engine %s", name)
	}
	return nil
}

// CreateDocuments https://www.elastic.co/guide/en/app-search/current/documents.html#documents-create
// max of 100 documents per batch, 100kb each document
func (d *elasticAppSearchService) CreateDocuments(docs []DeviceDefinitionSearchDoc, engineName string) ([]CreateDocsResp, error) {
	// todo: make docs generic parameter?
	if len(docs) > 100 {
		return nil, fmt.Errorf("docs slice is too big with len: %d, max of 100 items allowed", len(docs))
	}
	path := fmt.Sprintf("/api/as/v1/engines/%s/documents", engineName)

	var resp []CreateDocsResp
	err := d.buildAndExecRequest("POST", path, docs, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating documents in engine: %s", engineName)
	}
	// todo: what about iterating over the resp errors property to log that?
	return resp, nil
}

// CreateDocumentsBatched splits chunks of 100 docs and calls CreateDocuments method with each chunk
func (d *elasticAppSearchService) CreateDocumentsBatched(docs []DeviceDefinitionSearchDoc, engineName string) error {
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
func (d *elasticAppSearchService) UpdateSearchSettingsForDeviceDefs(engineName string) error {
	path := fmt.Sprintf("/api/as/v1/engines/%s/search_settings", engineName)
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
	err := d.buildAndExecRequest("PUT", path, body, nil)
	if err != nil {
		return errors.Wrapf(err, "error when trying to update search_settings for: %s", engineName)
	}
	return nil
}

// buildAndExecRequest makes request with token and headers, marshalling passed in obj or if string just passing along in body,
// and unmarshalling response body to objOut (must be passed in as ptr eg &varName)
func (d *elasticAppSearchService) buildAndExecRequest(method, url string, obj interface{}, objOut interface{}) error {
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

// DeviceDefinitionSearchDoc used for elastic search document indexing. entirely for searching, source of truth is DB.
// elastic only supports lowercase letters, number and underscores, ie. snake_case
type DeviceDefinitionSearchDoc struct {
	ID string `json:"id"`
	// SearchDisplay M+M+Y
	SearchDisplay string                               `json:"search_display"`
	Make          string                               `json:"make"`
	Model         string                               `json:"model"`
	Year          int                                  `json:"year"`
	Type          string                               `json:"type"`
	Attributes    []DeviceDefinitionAttributeSearchDoc `json:"device_attributes"`
	// SubModels: M+M+Y+Submodel name
	SubModels []string `json:"sub_models"`
	ImageURL  string   `json:"image_url"`
	MakeSlug  string   `json:"make_slug"`
	ModelSlug string   `json:"model_slug"`
	// future: we might add styles
}

type DeviceDefinitionAttributeSearchDoc struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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
