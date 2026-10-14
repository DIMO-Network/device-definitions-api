package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search"
)

type GetAllDeviceDefinitionByAutocompleteQuery struct {
	Query string `json:"query"`
}

type GetAllDeviceDefinitionByAutocompleteQueryResult struct {
	Items []GetAllDeviceDefinitionAutocompleteItem `json:"items"`
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

	// firstHitPerGroup, not *result.Hits: Autocomplete groups on definition_id,
	// and Typesense returns a grouped search in grouped_hits with hits unset --
	// so the old dereference of result.Hits was both a nil panic and, ungrouped,
	// ten trims of the same definition.
	hits := firstHitPerGroup(result)
	deviceDefinitions := make([]GetAllDeviceDefinitionAutocompleteItem, 0, len(hits))
	for _, hit := range hits {
		if hit.Document == nil {
			continue
		}
		doc := *hit.Document

		// docString and definitionName rather than bare type assertions: the
		// documents are untyped maps, a missing key panics on assertion, and the
		// per-trim document's own name carries the trim ("2020 Toyota Camry LE")
		// while this endpoint lists definitions.
		id := docString(doc, "definition_id")
		if id == "" {
			continue
		}
		deviceDefinitions = append(deviceDefinitions, GetAllDeviceDefinitionAutocompleteItem{
			ID:   id,
			Name: definitionName(doc),
		})
	}

	response := &GetAllDeviceDefinitionByAutocompleteQueryResult{
		Items: deviceDefinitions,
	}

	if response.Items == nil {
		response.Items = []GetAllDeviceDefinitionAutocompleteItem{}
	}
	return response, nil
}
