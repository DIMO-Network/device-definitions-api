package elasticsearch

import (
	"fmt"

	"github.com/pkg/errors"
)

type ElasticFilterResult struct {
	DocCount int `json:"doc_count"`
}

type DeviceFeaturesResp struct {
	Aggregations struct {
		Integrations struct {
			Buckets []struct {
				Key               string `json:"key"`
				DeviceDefinitions struct {
					Buckets []struct {
						Key     string `json:"key"`
						Regions struct {
							Buckets []struct {
								Key      string `json:"key"`
								Features struct {
									Buckets map[string]ElasticFilterResult
								}
							}
						}
					}
				}
			}
		}
	}
}

func (d *ElasticSearch) GetDeviceFeatures(envName, filterList string) (DeviceFeaturesResp, error) {
	url := fmt.Sprintf("%s/device-status-%s-*/_search", d.BaseURL, envName)
	body := `
{
	"size": 0,
	"query": {
		"bool": {
			"must": [
				{
					"exists": {
						"field": "data.deviceDefinitionId"
					}
				},
				{
					"exists": {
						"field": "data.region"
					}
				}
			]
		}
	},
	"aggs": {
		"integrations": {
			"terms": {
				"field": "source",
				"size": 1000
			},
			"aggs": {
				"deviceDefinitions": {
					"terms": {
						"field": "data.deviceDefinitionId",
						"size": 10000
					},
					"aggs": {
						"regions": {
							"terms": {
								"field": "data.region",
								"size": 1000
							},
							"aggs": {
								"features": {
									"filters": {
										"filters": ` + filterList + `
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
`

	deviceFeatures := DeviceFeaturesResp{}

	err := d.buildAndExecRequest("POST", url, body, &deviceFeatures)
	if err != nil {
		return DeviceFeaturesResp{}, errors.Wrap(err, "error when trying to fetch device features from elasticsearch")
	}

	return deviceFeatures, nil
}
