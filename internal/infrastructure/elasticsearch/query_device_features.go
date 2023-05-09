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

type deviceFeaturesResp2 struct {
	Aggregations struct {
		MyBuckets struct {
			Meta struct {
			} `json:"meta"`
			AfterKey struct {
				DataRegion             string `json:"data.region"`
				DataDeviceDefinitionID string `json:"data.deviceDefinitionId"`
			} `json:"after_key"`
			Buckets []struct {
				Key struct {
					DataRegion             string `json:"data.region"`
					DataDeviceDefinitionID string `json:"data.deviceDefinitionId"`
				} `json:"key"`
				DocCount     int `json:"doc_count"`
				Integrations struct {
					DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
					SumOtherDocCount        int `json:"sum_other_doc_count"`
					Buckets                 []struct {
						Key               string `json:"key"`
						DocCount          int    `json:"doc_count"`
						DeviceDefinitions struct {
							DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
							SumOtherDocCount        int `json:"sum_other_doc_count"`
							Buckets                 []struct {
								Key         []string `json:"key"`
								KeyAsString string   `json:"key_as_string"`
								DocCount    int      `json:"doc_count"`
								Regions     struct {
									DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
									SumOtherDocCount        int `json:"sum_other_doc_count"`
									Buckets                 []struct {
										Key      string `json:"key"`
										DocCount int    `json:"doc_count"`
										Features struct {
											Buckets map[string]ElasticFilterResult `json:"buckets"`
										} `json:"features"`
									} `json:"buckets"`
								} `json:"regions"`
							} `json:"buckets"`
						} `json:"deviceDefinitions"`
					} `json:"buckets"`
				} `json:"integrations,omitempty"`
			} `json:"buckets"`
		} `json:"my_buckets"`
	} `json:"aggregations"`
}

type bucketResp struct {
	Key struct {
		DataRegion             string `json:"data.region"`
		DataDeviceDefinitionID string `json:"data.deviceDefinitionId"`
	} `json:"key"`
	DocCount     int `json:"doc_count"`
	Integrations struct {
		DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int `json:"sum_other_doc_count"`
		Buckets                 []struct {
			Key               string `json:"key"`
			DocCount          int    `json:"doc_count"`
			DeviceDefinitions struct {
				DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
				SumOtherDocCount        int `json:"sum_other_doc_count"`
				Buckets                 []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
					Regions  struct {
						DocCountErrorUpperBound int `json:"doc_count_error_upper_bound"`
						SumOtherDocCount        int `json:"sum_other_doc_count"`
						Buckets                 []struct {
							Key      string `json:"key"`
							DocCount int    `json:"doc_count"`
							Features struct {
								Buckets map[string]interface{} `json:"buckets"`
							} `json:"features"`
						} `json:"buckets"`
					} `json:"regions"`
				} `json:"buckets"`
			} `json:"deviceDefinitions"`
		} `json:"buckets"`
	} `json:"integrations"`
}

// GetDeviceFeatures queries elastic for the presence of the given filterList (integration_features), returns all of them regardless if seen or not
func (d *ElasticSearch) GetDeviceFeatures(envName string, filterList map[string]any) (deviceFeaturesResp2, error) {
	var aggregatedResponses deviceFeaturesResp2
	var request getDeviceFeaturesRequest

	request.Size = 0
	request.Sort = []map[string]string{
		{
			"data.deviceDefinitionId": "asc",
			"data.region":             "asc",
		},
	}
	deviceDefIDExists := existsFilter{}
	deviceDefIDExists.Exists.Field = "data.deviceDefinitionId"
	request.Query.Bool.Must = append(request.Query.Bool.Must, deviceDefIDExists)

	regionExists := existsFilter{}
	regionExists.Exists.Field = "data.region"
	request.Query.Bool.Must = append(request.Query.Bool.Must, regionExists)

	request.Aggs.MyBuckets.Composite.Size = 100
	deviceDefIDSource := deviceDefSource{}
	deviceDefIDSource.DataDeviceDefinitionID.Terms.Field = "data.deviceDefinitionId"
	regionSource := regionSource{}
	regionSource.DataRegion.Terms.Field = "data.region"

	request.Aggs.MyBuckets.Composite.Sources = append(request.Aggs.MyBuckets.Composite.Sources, regionSource)
	request.Aggs.MyBuckets.Composite.Sources = append(request.Aggs.MyBuckets.Composite.Sources, deviceDefIDSource)

	request.Aggs.MyBuckets.Aggs.Integrations.Terms.Field = "source"
	request.Aggs.MyBuckets.Aggs.Integrations.Terms.Size = 10000
	request.Aggs.MyBuckets.Aggs.Integrations.Aggs.DeviceDefinitions.MultiTerms.Terms = []field{
		{Field: "data.deviceDefinitionId"}, {Field: "data.makeSlug.keyword"}, {Field: "data.modelSlug.keyword"},
	}
	request.Aggs.MyBuckets.Aggs.Integrations.Aggs.DeviceDefinitions.MultiTerms.Size = 10000
	request.Aggs.MyBuckets.Aggs.Integrations.Aggs.DeviceDefinitions.Aggs.Regions.Terms.Field = "data.region"
	request.Aggs.MyBuckets.Aggs.Integrations.Aggs.DeviceDefinitions.Aggs.Regions.Terms.Size = 10000
	request.Aggs.MyBuckets.Aggs.Integrations.Aggs.DeviceDefinitions.Aggs.Regions.Aggs.Features.Filters.Filters = filterList

	url := fmt.Sprintf("%s/device-status-%s-*/_search", d.BaseURL, envName)
	for {
		deviceFeatures := deviceFeaturesResp2{}
		err := d.buildAndExecRequest("POST", url, request, &deviceFeatures)
		if err != nil {
			return deviceFeaturesResp2{}, errors.Wrap(err, "error when trying to fetch device features from elasticsearch")
		}
		aggregatedResponses.Aggregations.MyBuckets.Buckets = append(aggregatedResponses.Aggregations.MyBuckets.Buckets, deviceFeatures.Aggregations.MyBuckets.Buckets...)
		if deviceFeatures.Aggregations.MyBuckets.AfterKey.DataDeviceDefinitionID == "" && deviceFeatures.Aggregations.MyBuckets.AfterKey.DataRegion == "" {
			break
		}
		request.Aggs.MyBuckets.Composite.After = map[string]string{
			"data.deviceDefinitionId": deviceFeatures.Aggregations.MyBuckets.AfterKey.DataDeviceDefinitionID,
			"data.region":             deviceFeatures.Aggregations.MyBuckets.AfterKey.DataRegion,
		}
	}

	return aggregatedResponses, nil
}

type getDeviceFeaturesRequest struct {
	Aggs struct {
		MyBuckets struct {
			Composite struct {
				Size    int               `json:"size,omitempty"`
				Sources []interface{}     `json:"sources,omitempty"`
				After   map[string]string `json:"after,omitempty"`
			} `json:"composite"`
			Aggs struct {
				Integrations struct {
					Terms struct {
						Field string `json:"field"`
						Size  int    `json:"size"`
					} `json:"terms"`
					Aggs struct {
						DeviceDefinitions struct {
							MultiTerms struct {
								Terms []field `json:"terms"`
								Size  int     `json:"size"`
							} `json:"multi_terms"`
							Aggs struct {
								Regions struct {
									Terms struct {
										Field string `json:"field"`
										Size  int    `json:"size"`
									} `json:"terms"`
									Aggs struct {
										Features struct {
											Filters struct {
												Filters map[string]interface{} `json:"filters"`
											} `json:"filters"`
										} `json:"features"`
									} `json:"aggs"`
								} `json:"regions"`
							} `json:"aggs"`
						} `json:"deviceDefinitions"`
					} `json:"aggs"`
				} `json:"integrations"`
			} `json:"aggs"`
		} `json:"my_buckets"`
	} `json:"aggs"`
	Size  int `json:"size"`
	Query struct {
		Bool struct {
			Must []existsFilter `json:"must"`
		} `json:"bool"`
	} `json:"query"`
	Sort []map[string]string `json:"sort"`
}

type existsFilter struct {
	Exists struct {
		Field string `json:"field"`
	} `json:"exists"`
}
type deviceDefSource struct {
	DataDeviceDefinitionID struct {
		Terms struct {
			Field string `json:"field"`
		} `json:"terms"`
	} `json:"data.deviceDefinitionId"`
}

type regionSource struct {
	DataRegion struct {
		Terms struct {
			Field string `json:"field"`
		} `json:"terms"`
	} `json:"data.region"`
}

type field struct {
	Field string `json:"field"`
}
