package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDevicesMMYQuery struct {
}

func (*GetDevicesMMYQuery) Key() string { return "GetDevicesMMYQuery" }

type GetDevicesMMYQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

type GetDevicesMMYQueryResult struct {
	Make  string `json:"make_slug"`
	Model string `json:"model_slug"`
	Year  int32  `json:"year"`
}

func NewGetDevicesMMYQueryHandler(repository repositories.DeviceDefinitionRepository) GetDevicesMMYQueryHandler {
	return GetDevicesMMYQueryHandler{
		Repository: repository,
	}
}

func (ch GetDevicesMMYQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	all, err := ch.Repository.GetDevicesMMY(ctx)
	if err != nil {
		return nil, err
	}
	result := &grpc.GetDevicesMMYResponse{}
	for _, v := range all {
		result.Device = append(result.Device, &grpc.GetDevicesMMYItemResponse{
			Make:  v.DeviceMakes.NameSlug,
			Model: v.DeviceDefinitions.ModelSlug,
			Year:  int32(v.DeviceDefinitions.Year),
		})

	}

	return result, nil
}
