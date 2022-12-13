package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetAllDevicesMakeModelYearQuery struct {
}

func (*GetAllDevicesMakeModelYearQuery) Key() string { return "GetAllDeviceMakeModelYearQuery" }

type GetAllDevicesMakeModelYearQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetAllDevicesMakeModelYearQueryHandler(repository repositories.DeviceDefinitionRepository) GetAllDevicesMakeModelYearQueryHandler {
	return GetAllDevicesMakeModelYearQueryHandler{
		Repository: repository,
	}
}

func (ch GetAllDevicesMakeModelYearQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, err := ch.Repository.GetAllDevicesMMY(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]interface{}, 0)

	for _, v := range all {
		deviceDef := v.DeviceDefinitions
		deviceMake := v.DeviceMakes
		resp := &grpc.DeviceType{
			Make:  deviceMake.NameSlug,
			Model: deviceDef.ModelSlug,
			Year:  int32(deviceDef.Year),
		}
		response = append(response, resp)
	}

	return response, nil
}
