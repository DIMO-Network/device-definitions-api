package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/common"
	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetByIdQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

type GetByIdQueryResult struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	Model              string `json:"model"`
	Year               int16  `json:"year"`
}

func (*GetByIdQuery) Key() string { return "GetByIdQuery" }

type GetByIdQueryHandler struct {
	Repository interfaces.IDeviceDefinitionRepository
}

func NewGetByIdQueryHandler(repository interfaces.IDeviceDefinitionRepository) GetByIdQueryHandler {
	return GetByIdQueryHandler{
		Repository: repository,
	}
}

func (ch GetByIdQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetByIdQuery)

	dd, _ := ch.Repository.GetById(ctx, qry.DeviceDefinitionID)

	if dd == nil {
		return nil, &common.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	return &GetByIdQueryResult{
		DeviceDefinitionID: dd.ID,
		Model:              dd.Model,
		Year:               dd.Year,
	}, nil
}
