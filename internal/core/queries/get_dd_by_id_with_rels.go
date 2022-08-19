package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/null/v8"
)

type GetDeviceDefinitionWithRelsQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetDeviceDefinitionWithRelsQueryResult struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

func (*GetDeviceDefinitionWithRelsQuery) Key() string {
	return "GetDeviceDefinitionWithRelsQuery"
}

type GetDeviceDefinitionWithRelsQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionWithRelsQueryHandler(repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionWithRelsQueryHandler {
	return GetDeviceDefinitionWithRelsQueryHandler{
		Repository: repository,
	}
}

func (ch GetDeviceDefinitionWithRelsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionWithRelsQuery)

	dd, _ := ch.Repository.GetWithIntegrations(ctx, qry.DeviceDefinitionID)

	if dd == nil {
		return nil, &common.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	// build object for integrations that have all the info
	var integrations []GetDeviceDefinitionWithRelsQueryResult
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			integrations = append(integrations, GetDeviceDefinitionWithRelsQueryResult{
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
