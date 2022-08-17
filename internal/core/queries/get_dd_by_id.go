package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByIdQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetDeviceDefinitionByIdQueryResult struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	Model              string `json:"model"`
	Year               int16  `json:"year"`
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

	return &GetDeviceDefinitionByIdQueryResult{
		DeviceDefinitionID: dd.ID,
		Model:              dd.Model,
		Year:               dd.Year,
	}, nil
}
