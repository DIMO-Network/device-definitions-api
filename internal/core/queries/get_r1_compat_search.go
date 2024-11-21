package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search"
	"github.com/mitchellh/mapstructure"
)

type GetR1CompatibilitySearch struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type GetR1CompatibilitySearchQueryResult struct {
	DeviceDefinitions []GetR1SearchEntryItem           `json:"deviceDefinitions"`
	Pagination        GetAllDeviceDefinitionPagination `json:"pagination"`
}

type GetR1SearchEntryItem struct {
	DefinitionID string `json:"definitionId"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Compatible   string `json:"compatible"`
	Name         string `json:"name"`
}

func (*GetR1CompatibilitySearch) Key() string {
	return "GetR1CompatibilitySearch"
}

type GetR1CompatibilitySearchQueryHandler struct {
	Service search.TypesenseAPIService
}

func NewGetR1CompatibilitySearchQueryHandler(service search.TypesenseAPIService) GetR1CompatibilitySearchQueryHandler {
	return GetR1CompatibilitySearchQueryHandler{
		Service: service,
	}
}

func (ch GetR1CompatibilitySearchQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetR1CompatibilitySearch)

	result, err := ch.Service.SearchR1Compatibility(ctx, qry.Query, qry.Page, qry.PageSize)
	if err != nil {
		return nil, err
	}

	deviceDefinitions := make([]GetR1SearchEntryItem, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		var doc map[string]interface{}
		if err := mapstructure.Decode(hit.Document, &doc); err != nil {
			continue
		}
		item := GetR1SearchEntryItem{
			DefinitionID: doc["definition_id"].(string),
			Make:         doc["make"].(string),
			Name:         doc["name"].(string),
			Model:        doc["model"].(string),
			Year:         int(doc["year"].(float64)),
			Compatible:   doc["compatible"].(string),
		}
		deviceDefinitions = append(deviceDefinitions, item)
	}

	pagination := GetAllDeviceDefinitionPagination{
		Page:       qry.Page,
		PageSize:   qry.PageSize,
		TotalItems: *result.Found,
		TotalPages: (*result.Found + qry.PageSize - 1) / qry.PageSize,
	}

	response := &GetR1CompatibilitySearchQueryResult{
		DeviceDefinitions: deviceDefinitions,
		Pagination:        pagination,
	}

	if response.DeviceDefinitions == nil {
		response.DeviceDefinitions = []GetR1SearchEntryItem{}
	}

	return response, nil
}
