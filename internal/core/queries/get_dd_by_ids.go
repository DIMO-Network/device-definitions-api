package queries

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByIdsQuery struct {
	DeviceDefinitionID []string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionByIdsQuery) Key() string { return "GetDeviceDefinitionByIdsQuery" }

type GetDeviceDefinitionByIdsQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionByIdsQueryHandler(repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionByIdsQueryHandler {
	return GetDeviceDefinitionByIdsQueryHandler{
		Repository: repository,
	}
}

func (ch GetDeviceDefinitionByIdsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIdsQuery)

	response := &grpc.GetDeviceDefinitionResponse{}

	for _, v := range qry.DeviceDefinitionID {
		dd, _ := ch.Repository.GetByID(ctx, v)

		if dd == nil {
			fmt.Printf("Not found")
			continue
		}

		rp := &grpc.GetDeviceDefinitionItemResponse{
			DeviceDefinitionId:     dd.ID,
			Name:                   fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
			ImageUrl:               dd.ImageURL.String,
			CompatibleIntegrations: []*grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations{},
			Type: &grpc.GetDeviceDefinitionItemResponse_Type{
				Type:  "Vehicle",
				Make:  dd.R.DeviceMake.Name,
				Model: dd.Model,
				Year:  int32(dd.Year),
			},
			Metadata: string(dd.Metadata.JSON),
			Verified: dd.Verified,
		}

		// vehicle info
		var vi map[string]GetDeviceVehicleInfo

		if err := dd.Metadata.Unmarshal(&vi); err == nil {

			numberOfDoors, _ := strconv.ParseInt(vi["vehicle_info"].NumberOfDoors, 6, 12)
			mpgHighway, _ := strconv.ParseFloat(vi["vehicle_info"].MPGHighway, 32)
			mpgCity, _ := strconv.ParseFloat(vi["vehicle_info"].MPGCity, 32)
			fuelTankCapacityGal, _ := strconv.ParseFloat(vi["vehicle_info"].FuelTankCapacityGal, 32)

			rp.VehicleData = &grpc.VehicleInfo{
				FuelType:            vi["vehicle_info"].FuelType,
				DrivenWheels:        vi["vehicle_info"].DrivenWheels,
				NumberOfDoors:       int32(numberOfDoors),
				Base_MSRP:           int32(vi["vehicle_info"].BaseMSRP),
				EPAClass:            vi["vehicle_info"].EPAClass,
				VehicleType:         vi["vehicle_info"].VehicleType,
				MPGHighway:          float32(mpgHighway),
				MPGCity:             float32(mpgCity),
				FuelTankCapacityGal: float32(fuelTankCapacityGal),
			}
		}

		if dd.R != nil {
			// compatible integrations
			rp.CompatibleIntegrations = buildDeviceCompatibility(dd.R.DeviceIntegrations)
			// sub_models
			rp.Type.SubModels = common.SubModelsFromStylesDB(dd.R.DeviceStyles)
		}

		// build object for integrations that have all the info
		rp.DeviceIntegrations = []*grpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{}
		if dd.R != nil {
			for _, di := range dd.R.DeviceIntegrations {
				rp.DeviceIntegrations = append(rp.DeviceIntegrations, &grpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
					Id:           di.R.Integration.ID,
					Type:         di.R.Integration.Type,
					Style:        di.R.Integration.Style,
					Vendor:       di.R.Integration.Vendor,
					Region:       di.Region,
					Capabilities: string(common.JSONOrDefault(di.Capabilities)),
				})
			}
		}

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}

// DeviceCompatibilityFromDB returns list of compatibility representation from device integrations db slice, assumes integration relation loaded
func buildDeviceCompatibility(dbDIS models.DeviceIntegrationSlice) []*grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations {
	if len(dbDIS) == 0 {
		return []*grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations{}
	}
	compatibilities := make([]*grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations, len(dbDIS))
	for i, di := range dbDIS {
		compatibilities[i] = &grpc.GetDeviceDefinitionItemResponse_CompatibleIntegrations{
			Id:           di.IntegrationID,
			Type:         di.R.Integration.Type,
			Style:        di.R.Integration.Style,
			Vendor:       di.R.Integration.Vendor,
			Region:       di.Region,
			Capabilities: string(common.JSONOrDefault(di.Capabilities)),
		}
	}
	return compatibilities
}
