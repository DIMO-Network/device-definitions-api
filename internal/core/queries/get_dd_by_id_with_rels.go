package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionWithRelsQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetDeviceDefinitionWithRelsQueryResult struct {
	ID                  string          `json:"id"`
	Type                string          `json:"type"`
	Style               string          `json:"style"`
	Vendor              string          `json:"vendor"`
	Region              string          `json:"region"`
	Capabilities        json.RawMessage `json:"capabilities"`
	Points              *int64          `json:"points"`
	ManufacturerTokenID *uint64         `json:"manufacturer_token_id"`
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
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	// build object for integrations that have all the info
	var integrations []GetDeviceDefinitionWithRelsQueryResult
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			points := int64(di.R.Integration.Points.Int)
			manufTknID, _ := di.R.Integration.ManufacturerTokenID.Uint64()
			integrations = append(integrations, GetDeviceDefinitionWithRelsQueryResult{
				ID:                  di.R.Integration.ID,
				Type:                di.R.Integration.Type,
				Style:               di.R.Integration.Style,
				Vendor:              di.R.Integration.Vendor,
				Region:              di.Region,
				Capabilities:        common.JSONOrDefault(di.Features),
				Points:              &points,
				ManufacturerTokenID: &manufTknID,
			})
		}
	}

	return integrations, nil
}
