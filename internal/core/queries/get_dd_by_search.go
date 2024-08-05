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
	PageSize int    `json:"pageSize"`
}

type GetAllDeviceDefinitionBySearchQueryResult struct {
	DeviceDefinitions []GetAllDeviceDefinitionItem     `json:"deviceDefinitions"`
	Facets            GetAllDeviceDefinitionFacet      `json:"facets"`
	Pagination        GetAllDeviceDefinitionPagination `json:"pagination"`
}

type GetAllDeviceDefinitionItem struct {
	ID                  string `json:"id"`
	DeviceDefinitionID  string `json:"legacy_ksuid"` //nolint
	Name                string `json:"name"`
	Make                string `json:"make"`
	ManufacturerTokenID int    `json:"manufacturerTokenId"`
	Model               string `json:"model"`
	Year                int    `json:"year"`
	ImageURL            string `json:"imageUrl"`
}

type GetAllDeviceDefinitionFacet struct {
	Makes  []GetAllDeviceDefinitionFacetItem `json:"makes"`
	Models []GetAllDeviceDefinitionFacetItem `json:"models"`
	Years  []GetAllDeviceDefinitionFacetItem `json:"years"`
}

type GetAllDeviceDefinitionFacetItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type GetAllDeviceDefinitionPagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
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

	deviceDefinitions := make([]GetAllDeviceDefinitionItem, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		var doc map[string]interface{}
		if err := mapstructure.Decode(hit.Document, &doc); err != nil {
			continue
		}
		item := GetAllDeviceDefinitionItem{
			ID:                  doc["id"].(string),
			DeviceDefinitionID:  doc["device_definition_id"].(string),
			Name:                doc["name"].(string),
			Make:                doc["make"].(string),
			ManufacturerTokenID: int(doc["manufacturer_token_id"].(float64)),
			Model:               doc["model"].(string),
			Year:                int(doc["year"].(float64)),
			ImageURL:            doc["image_url"].(string),
		}
		deviceDefinitions = append(deviceDefinitions, item)
	}

	var makes []GetAllDeviceDefinitionFacetItem
	var models []GetAllDeviceDefinitionFacetItem
	var years []GetAllDeviceDefinitionFacetItem

	for _, facet := range *result.FacetCounts {
		for _, count := range *facet.Counts {
			if *facet.FieldName == "make" {
				makes = append(makes, GetAllDeviceDefinitionFacetItem{
					Name:  *count.Value,
					Count: *count.Count,
				})
			}
			if *facet.FieldName == "model" {
				models = append(models, GetAllDeviceDefinitionFacetItem{
					Name:  *count.Value,
					Count: *count.Count,
				})
			}
			if *facet.FieldName == "year" {
				years = append(years, GetAllDeviceDefinitionFacetItem{
					Name:  *count.Value,
					Count: *count.Count,
				})
			}
		}
	}

	facets := GetAllDeviceDefinitionFacet{
		Makes:  makes,
		Models: models,
		Years:  years,
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
	if response.Facets.Makes == nil {
		response.Facets.Makes = []GetAllDeviceDefinitionFacetItem{}
	}
	if response.Facets.Models == nil {
		response.Facets.Models = []GetAllDeviceDefinitionFacetItem{}
	}
	if response.Facets.Years == nil {
		response.Facets.Years = []GetAllDeviceDefinitionFacetItem{}
	}

	return response, nil
}
