package queries

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search"
	"github.com/mitchellh/mapstructure"
)

type GetAllDeviceDefinitionByAutocompleteQuery struct {
	Query string `json:"query"`
}

type GetAllDeviceDefinitionByAutocompleteQueryResult struct {
	DeviceDefinitions []GetAllDeviceDefinitionAutocompleteItem `json:"items"`
}

type GetAllDeviceDefinitionAutocompleteItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (*GetAllDeviceDefinitionByAutocompleteQuery) Key() string {
	return "GetAllDeviceDefinitionByAutocompleteQuery"
}

type GetAllDeviceDefinitionByAutocompleteQueryHandler struct {
	Service search.TypesenseAPIService
}

func NewGetAllDeviceDefinitionByAutocompleteQueryHandler(service search.TypesenseAPIService) GetAllDeviceDefinitionByAutocompleteQueryHandler {
	return GetAllDeviceDefinitionByAutocompleteQueryHandler{
		Service: service,
	}
}

func (ch GetAllDeviceDefinitionByAutocompleteQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetAllDeviceDefinitionByAutocompleteQuery)

	result, err := ch.Service.Autocomplete(ctx, qry.Query)
	if err != nil {
		return nil, err
	}

	deviceDefinitions := make([]GetAllDeviceDefinitionAutocompleteItem, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		var doc map[string]interface{}
		if err := mapstructure.Decode(hit.Document, &doc); err != nil {
			continue
		}

		item := GetAllDeviceDefinitionAutocompleteItem{
			ID:   doc["id"].(string),
			Name: doc["name"].(string),
		}
		deviceDefinitions = append(deviceDefinitions, item)
	}

	response := &GetAllDeviceDefinitionByAutocompleteQueryResult{
		DeviceDefinitions: deviceDefinitions,
	}

	if response.DeviceDefinitions == nil {
		response.DeviceDefinitions = []GetAllDeviceDefinitionAutocompleteItem{}
	}
	return response, nil
}
