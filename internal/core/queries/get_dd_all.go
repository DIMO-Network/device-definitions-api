package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetAllDeviceDefinitionQuery struct {
}

func (*GetAllDeviceDefinitionQuery) Key() string { return "GetAllDeviceDefinitionQuery" }

type GetAllDeviceDefinitionQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetAllDeviceDefinitionQueryHandler(repository repositories.DeviceDefinitionRepository) GetAllDeviceDefinitionQueryHandler {
	return GetAllDeviceDefinitionQueryHandler{
		Repository: repository,
	}
}

func (ch GetAllDeviceDefinitionQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, _ := ch.Repository.GetAll(ctx, true)

	response := &grpc.GetDeviceDefinitionResponse{}
	for _, v := range all {
		dd := common.BuildFromDeviceDefinitionToQueryResult(v)
		rp := common.BuildFromQueryResultToGRPC(dd)

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
