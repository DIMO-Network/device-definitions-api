package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByIdQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetDeviceDefinitionByIdQueryResult struct {
	DeviceDefinitionID     string                               `json:"deviceDefinitionId"`
	Name                   string                               `json:"name"`
	ImageURL               *string                              `json:"imageUrl"`
	CompatibleIntegrations []GetDeviceCompatibility             `json:"compatibleIntegrations"`
	Type                   DeviceType                           `json:"type"`
	VehicleInfo            GetDeviceVehicleInfo                 `json:"vehicleData,omitempty"`
	Metadata               interface{}                          `json:"metadata"`
	Verified               bool                                 `json:"verified"`
	DeviceIntegrations     []GetDeviceDefinitionIntegrationList `json:"deviceIntegrations"`
}

type GetDeviceDefinitionIntegrationList struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Style        string          `json:"style"`
	Vendor       string          `json:"vendor"`
	Region       string          `json:"region"`
	Country      string          `json:"country,omitempty"`
	Capabilities json.RawMessage `json:"capabilities"`
}

func (*GetDeviceDefinitionByIdQuery) Key() string { return "GetDeviceDefinitionByIdQuery" }

type GetDeviceDefinitionByIdQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionByIdQueryHandler(repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionByIdQueryHandler {
	return GetDeviceDefinitionByIdQueryHandler{
		Repository: repository,
	}
}

func (ch GetDeviceDefinitionByIdQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIdQuery)

	dd, _ := ch.Repository.GetById(ctx, qry.DeviceDefinitionID)

	if dd == nil {
		return nil, &common.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	rp := GetDeviceDefinitionByIdQueryResult{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.Ptr(),
		CompatibleIntegrations: []GetDeviceCompatibility{},
		Type: DeviceType{
			Type:  "Vehicle",
			Make:  dd.R.DeviceMake.Name,
			Model: dd.Model,
			Year:  int(dd.Year),
		},
		Metadata: string(dd.Metadata.JSON),
		Verified: dd.Verified,
	}

	// vehicle info
	var vi map[string]GetDeviceVehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		rp.VehicleInfo = vi["vehicle_info"]
	}

	if dd.R != nil {
		// compatible integrations
		rp.CompatibleIntegrations = deviceCompatibilityFromDB(dd.R.DeviceIntegrations)
		// sub_models
		rp.Type.SubModels = subModelsFromStylesDB(dd.R.DeviceStyles)
	}

	// build object for integrations that have all the info
	rp.DeviceIntegrations = []GetDeviceDefinitionIntegrationList{}
	if dd.R != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.DeviceIntegrations = append(rp.DeviceIntegrations, GetDeviceDefinitionIntegrationList{

				ID:           di.R.Integration.ID,
				Type:         di.R.Integration.Type,
				Style:        di.R.Integration.Style,
				Vendor:       di.R.Integration.Vendor,
				Region:       di.Region,
				Capabilities: jsonOrDefault(di.Capabilities),
			})
		}
	}

	return rp, nil
}
