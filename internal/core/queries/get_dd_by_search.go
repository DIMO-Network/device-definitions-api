package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/search"
	"github.com/typesense/typesense-go/typesense/api"
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
	ID                 string `json:"id"`
	DeviceDefinitionID string `json:"legacy_ksuid"` //nolint
	Name               string `json:"name"`
	Make               string `json:"make"`
	// ManufacturerTokenID int    `json:"manufacturerTokenId"` // todo
	Model    string `json:"model"`
	Year     int    `json:"year"`
	ImageURL string `json:"imageUrl"`
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

	hits := firstHitPerGroup(result)
	deviceDefinitions := make([]GetAllDeviceDefinitionItem, 0, len(hits))
	for _, hit := range hits {
		if hit.Document == nil {
			continue
		}
		doc := *hit.Document
		id := docString(doc, "definition_id")
		item := GetAllDeviceDefinitionItem{
			ID: id,
			// Legacy ksuids no longer exist and the index carries no
			// device_definition_id. The field stays on the wire for
			// compatibility and carries the slug id, the only id there is.
			DeviceDefinitionID: id,
			Name:               definitionName(doc),
			Make:               docString(doc, "make"),
			Model:              docString(doc, "model"),
			Year:               docInt(doc, "year"),
			ImageURL:           docString(doc, "image_url"),
		}
		deviceDefinitions = append(deviceDefinitions, item)
	}

	// Facet counts are per document and the index holds one document per
	// trim, so each count is trims, not definitions: Typesense does not count
	// groups.
	var makes []GetAllDeviceDefinitionFacetItem
	var models []GetAllDeviceDefinitionFacetItem
	var years []GetAllDeviceDefinitionFacetItem

	var facetCounts []api.FacetCounts
	if result.FacetCounts != nil {
		facetCounts = *result.FacetCounts
	}
	for _, facet := range facetCounts {
		if facet.Counts == nil || facet.FieldName == nil {
			continue
		}
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

	found := 0
	if result.Found != nil {
		found = *result.Found
	}
	pagination := GetAllDeviceDefinitionPagination{
		Page:       qry.Page,
		PageSize:   qry.PageSize,
		TotalItems: found,
		TotalPages: (found + qry.PageSize - 1) / qry.PageSize,
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

// firstHitPerGroup flattens a group_by=definition_id result to one hit per
// definition. The worker indexes one document per trim, so without grouping a
// query for "camry" returns toyota_camry_2020 once per trim; the endpoint's
// contract is one item per definition. Ungrouped hits are the fallback so a
// fake, or an index queried without group_by, still works.
func firstHitPerGroup(result *api.SearchResult) []api.SearchResultHit {
	if result.GroupedHits != nil {
		hits := make([]api.SearchResultHit, 0, len(*result.GroupedHits))
		for _, group := range *result.GroupedHits {
			if len(group.Hits) > 0 {
				hits = append(hits, group.Hits[0])
			}
		}
		return hits
	}
	if result.Hits != nil {
		return *result.Hits
	}
	return nil
}

// docString reads a string field from a Typesense document, "" when absent
// or not a string. Documents are untyped maps: a direct type assertion on a
// key the index does not carry panics, and fiber's recover turns that into a
// 500 on every hit.
func docString(doc map[string]interface{}, key string) string {
	s, _ := doc[key].(string)
	return s
}

// docInt reads a numeric field, 0 when absent. JSON numbers decode as float64.
func docInt(doc map[string]interface{}, key string) int {
	switch v := doc[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

// definitionName is the definition-level display name. Each document in the
// index is a trim and its name carries the trim ("2020 Toyota Camry LE"), but
// this endpoint answers one item per definition and ranking decides which
// trim's document leads the group. It is built the way definitions were always
// named, falling back to the document's name when a part is missing.
func definitionName(doc map[string]interface{}) string {
	year, mk, model := docInt(doc, "year"), docString(doc, "make"), docString(doc, "model")
	if year == 0 || mk == "" || model == "" {
		return docString(doc, "name")
	}
	return common.BuildDeviceDefinitionName(int16(year), mk, model)
}
