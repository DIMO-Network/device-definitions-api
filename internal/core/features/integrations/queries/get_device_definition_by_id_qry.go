package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/common"
	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/null/v8"
)

type GetByDeviceDefinitionIntegationIdQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetByDeviceDefinitionIntegrationIdQueryResult struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

func (*GetByDeviceDefinitionIntegationIdQuery) Key() string {
	return "GetByDeviceDefinitionIntegationIdQuery"
}

type GetByDeviceDefinitionIntegrationIdQueryHandler struct {
	Repository interfaces.IDeviceDefinitionRepository
}

func NewGetByDeviceDefinitionIntegrationIdQueryHandler(repository interfaces.IDeviceDefinitionRepository) GetByDeviceDefinitionIntegrationIdQueryHandler {
	return GetByDeviceDefinitionIntegrationIdQueryHandler{
		Repository: repository,
	}
}

func (ch GetByDeviceDefinitionIntegrationIdQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetByDeviceDefinitionIntegationIdQuery)

	dd, _ := ch.Repository.GetWithIntegrations(ctx, qry.DeviceDefinitionID)

	if dd == nil {
		return nil, &common.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	// build object for integrations that have all the info
	var integrations []GetByDeviceDefinitionIntegrationIdQueryResult
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			integrations = append(integrations, GetByDeviceDefinitionIntegrationIdQueryResult{
				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendor,
				Region:       di.Region,
				Capabilities: jsonOrDefault(di.Capabilities),
			})
		}
	}

	return integrations, nil
}

func jsonOrDefault(j null.JSON) json.RawMessage {
	if !j.Valid || len(j.JSON) == 0 {
		return []byte(`{}`)
	}
	return j.JSON
}
