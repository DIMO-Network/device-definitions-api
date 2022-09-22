package elasticsearch

import (
	"fmt"

	"github.com/pkg/errors"
)

type DeviceFeatures struct {
	BatteryVoltage struct {
		DocumentCount int64 `json:"doc_count"`
	}
	FuelPercentRemaining struct {
		DocumentCount int64 `json:"doc_count"`
	}
	Odometer struct {
		DocumentCount int64 `json:"doc_count"`
	}
	Oil struct {
		DocumentCount int64 `json:"doc_count"`
	}
	Soc struct {
		DocumentCount int64 `json:"doc_count"`
	}
	Speed struct {
		DocumentCount int64 `json:"doc_count"`
	}
	Tires struct {
		DocumentCount int64 `json:"doc_count"`
	}
}

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
						Key           string `json:"key"`
						DocumentCount int64  `json:"doc_count"`
						Features      struct {
							Buckets map[string]map[string]int
						}
					}
				}
			}
		}
	}
}

func (d *ElasticSearch) GetDeviceFeatures(envName string) (DeviceFeaturesResp, error) {
	url := fmt.Sprintf("%s/device-status-%s*/_search", d.BaseURL, envName)
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
			  }
			]
		  }
		},
		"aggs": {
		  "features": {
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
			  "features":{
				"filters": {
				  "filters": {
					"odometer":{
					  "exists": {
						"field": "data.odometer"
					  }
					},
					"fuelPercentRemaining": {
					  "exists": {
						  "field": "data.fuelPercentRemaining"
					  }
					},
					"oil": {
					  "exists": {
						  "field": "data.oil"
					  }
					},
					"soc": {
					  "exists": {
						  "field": "data.soc"
					  }
					},
					"speed": {
					  "exists": {
						  "field": "data.speed"
					  }
					},
					"tires": {
					  "exists": {
						  "field": "data.tires.frontLeft"
					  }
					},
					"batteryVoltage":{
					  "exists": {
						"field": "data.batteryVoltage"
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
	  }
	`

	deviceFeatures := DeviceFeaturesResp{}

	_, err := d.buildAndExecRequest("POST", url, body, &deviceFeatures)
	if err != nil {
		return DeviceFeaturesResp{}, errors.Wrap(err, "error when trying to fetch device features from elasticsearch")
	}

	return deviceFeatures, nil
}
