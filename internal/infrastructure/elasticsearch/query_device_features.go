package elasticsearch

import (
	"fmt"

	"github.com/pkg/errors"
)

type DeviceFeaturesResp struct {
	Aggregations struct {
		Features struct {
			SumOtherDocCount int64 `json:"sum_other_doc_count"`
			Buckets          []struct {
				Key               string `json:"key"`
				DocumentCount     int64  `json:"doc_count"`
				DeviceDefinitions struct {
					SumOtherDocCount int64 `json:"sum_other_doc_count"`
					Buckets          []struct {
						Key      string `json:"key"`
						DocCount int64  `json:"doc_count"`
						Features struct {
							Buckets map[string]map[string]int
						}
					}
				}
			}
		}
	}
}

func (d *ElasticSearch) GetDeviceFeatures(envName, fields string) (DeviceFeaturesResp, error) {
	url := fmt.Sprintf("%s/device-status-%s*/_search", d.BaseURL, envName)
	body := `
	{
		"size":0,
		"query":{
		   "bool":{
			  "must":[
				 {
					"exists":{
					   "field":"data.deviceDefinitionId"
					}
				 }
			  ]
		   }
		},
		"aggs":{
		   "features":{
			  "terms":{
				 "field":"source",
				 "size":1000
			  },
			  "aggs":{
				 "deviceDefinitions":{
					"terms":{
					   "field":"data.deviceDefinitionId",
					   "size":10000
					},
					"aggs":{
					   "features":{
						  "filters":{
							"filters":` + fields + `
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

	_, err := d.buildAndExecRequest("POST", url, body, &deviceFeatures)
	if err != nil {
		return DeviceFeaturesResp{}, errors.Wrap(err, "error when trying to fetch device features from elasticsearch")
	}

	return deviceFeatures, nil
}
