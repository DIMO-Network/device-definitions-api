package queries

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByMakeModelYearQuery struct {
	Make  string `json:"make" validate:"required"`
	Model string `json:"model" validate:"required"`
	Year  int    `json:"year" validate:"required"`
}

func (*GetDeviceDefinitionByMakeModelYearQuery) Key() string {
	return "GetDeviceDefinitionByMakeModelYearQuery"
}

type GetDeviceDefinitionByMakeModelYearQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionByMakeModelYearQueryHandler(repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionByMakeModelYearQueryHandler {
	return GetDeviceDefinitionByMakeModelYearQueryHandler{
		Repository: repository,
	}
}

func (ch GetDeviceDefinitionByMakeModelYearQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByMakeModelYearQuery)

	dd, _ := ch.Repository.GetByMakeModelAndYears(ctx, qry.Make, qry.Model, qry.Year, true)

	if dd == nil {
		return &GetDeviceDefinitionQueryResult{}, nil
	}

	rp := GetDeviceDefinitionQueryResult{
		DeviceDefinitionID:     dd.ID,
		Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:               dd.ImageURL.String,
		CompatibleIntegrations: []GetDeviceCompatibility{},
		DeviceMake: DeviceMake{
			ID:              dd.R.DeviceMake.ID,
			Name:            dd.R.DeviceMake.Name,
			LogoURL:         dd.R.DeviceMake.LogoURL,
			OemPlatformName: dd.R.DeviceMake.OemPlatformName,
		},
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
		rp.Type.SubModels = common.SubModelsFromStylesDB(dd.R.DeviceStyles)
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
				Capabilities: common.JSONOrDefault(di.Capabilities),
			})
		}
	}

	return rp, nil
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func deviceCompatibilityFromDB(dbDIS models.DeviceIntegrationSlice) []GetDeviceCompatibility {
	if len(dbDIS) == 0 {
		return []GetDeviceCompatibility{}
	}
	compatibilities := make([]GetDeviceCompatibility, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = GetDeviceCompatibility{
			ID:           di.IntegrationID,
			Type:         di.R.Integration.Type,
			Style:        di.R.Integration.Style,
			Vendor:       di.R.Integration.Vendor,
			Region:       di.Region,
			Capabilities: common.JSONOrDefault(di.Capabilities),
		}
	}
	return compatibilities
}
