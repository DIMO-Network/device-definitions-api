package queries

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search"
	"github.com/mitchellh/mapstructure"
)

type GetAllDeviceDefinitionBySearchQuery struct {
	Query    string `json:"query"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Year     int    `json:"year"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GetAllDeviceDefinitionBySearchQueryResult struct {
	DeviceDefinitions []GetAllDeviceDefinitionItem     `json:"device_definitions"`
	Facets            []GetAllDeviceDefinitionFacet    `json:"facets"`
	Pagination        GetAllDeviceDefinitionPagination `json:"pagination"`
}

type GetAllDeviceDefinitionItem struct {
	ID                 string `json:"id"`
	DeviceDefinitionID string `json:"legacy_ksuid"`
	Name               string `json:"name"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	Year               int    `json:"year"`
	ImageURL           string `json:"image_url"`
}

type GetAllDeviceDefinitionFacet struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type GetAllDeviceDefinitionPagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func (*GetAllDeviceDefinitionBySearchQuery) Key() string {
	return "GetAllDeviceDefinitionBySearchQuery"
}

type GetAllDeviceDefinitionBySearchQueryHandler struct {
	Service search.TypesenseAPIService
}

func NewGetAllDeviceDefinitionBySearchQueryHandler(service search.TypesenseAPIService) GetAllDeviceDefinitionBySearchQueryHandler {
	return GetAllDeviceDefinitionBySearchQueryHandler{
		Service: service,
	}
}

func (ch GetAllDeviceDefinitionBySearchQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetAllDeviceDefinitionBySearchQuery)

	result, err := ch.Service.GetDeviceDefinitions(ctx, qry.Query, qry.Make, qry.Model, qry.Year, qry.Page, qry.PageSize)
	if err != nil {
		return nil, err
	}

	var deviceDefinitions []GetAllDeviceDefinitionItem
	for _, hit := range *result.Hits {
		var doc map[string]interface{}
		if err := mapstructure.Decode(hit.Document, &doc); err != nil {
			continue
		}
		item := GetAllDeviceDefinitionItem{
			ID:                 doc["id"].(string),
			DeviceDefinitionID: doc["device_definition_id"].(string),
			Name:               doc["name"].(string),
			Make:               doc["make"].(string),
			Model:              doc["model"].(string),
			Year:               int(doc["year"].(float64)),
			ImageURL:           doc["image_url"].(string),
		}
		deviceDefinitions = append(deviceDefinitions, item)
	}

	var facets []GetAllDeviceDefinitionFacet
	for _, facet := range *result.FacetCounts {
		for _, count := range *facet.Counts {
			facets = append(facets, GetAllDeviceDefinitionFacet{
				Name:  *count.Value,
				Count: *count.Count,
			})
		}
	}

	pagination := GetAllDeviceDefinitionPagination{
		Page:       qry.Page,
		PageSize:   qry.PageSize,
		TotalItems: *result.Found,
		TotalPages: (*result.Found + qry.PageSize - 1) / qry.PageSize,
	}

	response := &GetAllDeviceDefinitionBySearchQueryResult{
		DeviceDefinitions: deviceDefinitions,
		Facets:            facets,
		Pagination:        pagination,
	}

	if response.DeviceDefinitions == nil {
		response.DeviceDefinitions = []GetAllDeviceDefinitionItem{}
	}
	if response.Facets == nil {
		response.Facets = []GetAllDeviceDefinitionFacet{}
	}

	return response, nil
}
