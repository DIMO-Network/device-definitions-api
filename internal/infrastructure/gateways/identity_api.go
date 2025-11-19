package gateways

import (
	"fmt"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/shared/pkg/http"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var ErrNotFound = errors.New("not found")
var ErrBadRequest = errors.New("bad request")

type identityAPIService struct {
	httpClient     http.ClientWrapper
	logger         zerolog.Logger
	identityAPIURL string
}

//go:generate mockgen -source identity_api.go -destination mocks/identity_api_mock.go -package mocks
type IdentityAPI interface {
	GetManufacturer(slug string) (*coremodels.Manufacturer, error)
	GetManufacturers() ([]coremodels.Manufacturer, error)
	GetDeviceDefinitionByID(id string) (*coremodels.DeviceDefinition, error)
}

// NewIdentityAPIService creates a new instance of IdentityAPI, initializing it with the provided logger, settings, and HTTP client.
// httpClient is used for testing really
func NewIdentityAPIService(logger *zerolog.Logger, settings *config.Settings) IdentityAPI {
	httpClient, _ := http.NewClientWrapper(settings.IdentityAPIURL.String(), "", 10*time.Second, nil, true) // ok to ignore err since only used for tor check

	return &identityAPIService{
		httpClient:     httpClient,
		logger:         *logger,
		identityAPIURL: settings.IdentityAPIURL.String(),
	}
}

// GetManufacturer from identity-api by the name - must match exactly. Returns the token id and other on chain info
func (i *identityAPIService) GetManufacturer(slug string) (*coremodels.Manufacturer, error) {
	query := `{
  manufacturer(by: {slug: "` + slug + `"}) {
    	tokenId
    	name
    	tableId
    	owner
  	  }
	}`
	var wrapper struct {
		Data struct {
			Manufacturer coremodels.Manufacturer `json:"manufacturer"`
		} `json:"data"`
	}
	err := i.httpClient.GraphQLQuery("", query, &wrapper)
	if err != nil {
		return nil, err
	}
	if wrapper.Data.Manufacturer.Name == "" {
		return nil, errors.Wrapf(ErrNotFound, "identity-api did not find manufacturer with slug: %s", slug)
	}
	return &wrapper.Data.Manufacturer, nil
}

func (i *identityAPIService) GetManufacturers() ([]coremodels.Manufacturer, error) {
	query := `{
  manufacturers {
    totalCount
    nodes {
      id
      tokenId
      name
      tableId
      owner
    }
  }
}`
	var wrapper struct {
		Data struct {
			Manufacturers struct {
				TotalCount int                       `json:"totalCount"`
				Nodes      []coremodels.Manufacturer `json:"nodes"`
			} `json:"manufacturers"`
		} `json:"data"`
	}

	err := i.httpClient.GraphQLQuery("", query, &wrapper)
	if err != nil {
		return nil, err
	}
	return wrapper.Data.Manufacturers.Nodes, nil
}

func (i *identityAPIService) GetDeviceDefinitionByID(id string) (*coremodels.DeviceDefinition, error) {
	graphqlQuery := fmt.Sprintf(DeviceDefinitionByIDQuery, id)

	var wrapper struct {
		Data struct {
			DeviceDefinition coremodels.DeviceDefinition `json:"deviceDefinition"`
		} `json:"data"`
	}

	err := i.httpClient.GraphQLQuery("", graphqlQuery, &wrapper)
	if err != nil {
		return nil, err
	}
	return &wrapper.Data.DeviceDefinition, nil
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

const DeviceDefinitionByIDQuery = `{
	deviceDefinition(by: {id: "%s"}) {
    	deviceDefinitionId
    	manufacturer {
      		name
      		tokenId
    	}
    	model
    	year
  	}
}`
