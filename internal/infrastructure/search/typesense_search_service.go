package search

import (
	"context"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/typesense/typesense-go/typesense"
)

type TypesenseAPIService interface {
	GetDeviceDefinitions(ctx context.Context, search string) error
}

type typesenseAPIService struct {
	settings *config.Settings
	log      *zerolog.Logger
	client   *typesense.Client
}

func NewTypesenseAPIService(settings *config.Settings, log *zerolog.Logger) TypesenseAPIService {
	client := typesense.NewClient(
		typesense.WithServer(settings.SearchServiceAPIURL),
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

func (t typesenseAPIService) GetDeviceDefinitions(ctx context.Context, search string) error {

	searchParameters := &api.SearchCollectionParams{
		Q:       search,
		QueryBy: "name",
		FacetBy: pointer.String("make,model,year"),
	}

	result, err := t.client.Collection(t.settings.SearchServiceIndexName).Documents().Search(ctx, searchParameters)

	if err != nil {
		return err
	}

	if result.FacetCounts != nil {

	}

	return nil
}
