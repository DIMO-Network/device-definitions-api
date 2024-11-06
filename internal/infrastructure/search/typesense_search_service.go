package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
)

//go:generate mockgen -source typesense_search_service.go -destination mocks/typesense_search_service_mock.go -package mocks
type TypesenseAPIService interface {
	GetDeviceDefinitions(ctx context.Context, search, make, model string, year, page, pageSize int) (*api.SearchResult, error)
	Autocomplete(ctx context.Context, search string) (*api.SearchResult, error)
	SearchR1Compatibility(ctx context.Context, search string, page, pageSize int) (*api.SearchResult, error)
}

type typesenseAPIService struct {
	settings *config.Settings
	log      *zerolog.Logger
	client   *typesense.Client
}

func NewTypesenseAPIService(settings *config.Settings, log *zerolog.Logger) TypesenseAPIService {
	client := typesense.NewClient(
		typesense.WithServer(settings.SearchServiceAPIURL.String()),
		typesense.WithAPIKey(settings.SearchServiceAPIKey),
		typesense.WithConnectionTimeout(5*time.Second),
		typesense.WithCircuitBreakerMaxRequests(50),
		typesense.WithCircuitBreakerInterval(2*time.Minute),
		typesense.WithCircuitBreakerTimeout(1*time.Minute))

	return &typesenseAPIService{
		settings: settings,
		log:      log,
		client:   client,
	}
}

func (t typesenseAPIService) GetDeviceDefinitions(ctx context.Context, search, make, model string, year, page, pageSize int) (*api.SearchResult, error) {

	var filters strings.Builder
	if make != "" {
		filters.WriteString(fmt.Sprintf("make_slug:=%s", make))
	}
	if model != "" {
		if filters.Len() > 0 {
			filters.WriteString(" && ")
		}
		filters.WriteString(fmt.Sprintf("model_slug:=%s", model))
	}
	if year != 0 {
		if filters.Len() > 0 {
			filters.WriteString(" && ")
		}
		filters.WriteString(fmt.Sprintf("year:=%d", year))
	}

	searchParameters := &api.SearchCollectionParams{
		Q:        search,
		QueryBy:  "name",
		FacetBy:  pointer.String("make,model,year"),
		Page:     pointer.Int(page),
		PerPage:  pointer.Int(pageSize),
		FilterBy: pointer.String(filters.String()),
	}

	result, err := t.client.Collection(t.settings.SearchServiceIndexName).Documents().Search(ctx, searchParameters)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t typesenseAPIService) Autocomplete(ctx context.Context, search string) (*api.SearchResult, error) {

	searchParameters := &api.SearchCollectionParams{
		Q:                       search,
		QueryBy:                 "name",
		Limit:                   pointer.Int(10),
		HighlightFullFields:     pointer.String("name"),
		HighlightAffixNumTokens: pointer.Int(2),
	}

	result, err := t.client.Collection(t.settings.SearchServiceIndexName).Documents().Search(ctx, searchParameters)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t typesenseAPIService) SearchR1Compatibility(ctx context.Context, search string, page, pageSize int) (*api.SearchResult, error) {
	searchParameters := &api.SearchCollectionParams{
		Q:       search,
		QueryBy: "model, make, year",
		FacetBy: pointer.String("make,model,year"),
		Page:    pointer.Int(page),
		PerPage: pointer.Int(pageSize),
	}

	result, err := t.client.Collection("r1_compatibility").Documents().Search(ctx, searchParameters)

	if err != nil {
		return nil, err
	}

	return result, nil
}
