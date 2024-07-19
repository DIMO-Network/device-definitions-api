package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
)

type GetAllDeviceDefinitionBySearchQuery struct {
}

func (*GetAllDeviceDefinitionBySearchQuery) Key() string {
	return "GetAllDeviceDefinitionBySearchQuery"
}

type GetAllDeviceDefinitionBySearchQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetAllDeviceDefinitionBySearchQueryHandler(repository repositories.DeviceDefinitionRepository) GetAllDeviceDefinitionBySearchQueryHandler {
	return GetAllDeviceDefinitionBySearchQueryHandler{
		Repository: repository,
	}
}

func (ch GetAllDeviceDefinitionBySearchQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {
	return nil, nil
}
