package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByIDQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionByIDQuery) Key() string { return "GetDeviceDefinitionByIdQuery" }

type GetDeviceDefinitionByIDQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceDefinitionByIDQueryHandler(cache services.DeviceDefinitionCacheService) GetDeviceDefinitionByIDQueryHandler {
	return GetDeviceDefinitionByIDQueryHandler{
		DDCache: cache,
	}
}

func (ch GetDeviceDefinitionByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDQuery)

	dd, err := ch.DDCache.GetDeviceDefinitionByID(ctx, qry.DeviceDefinitionID)

	if dd == nil {
		return nil, &common.NotFoundError{
			Err: err,
		}
	}

	return dd, nil
}
