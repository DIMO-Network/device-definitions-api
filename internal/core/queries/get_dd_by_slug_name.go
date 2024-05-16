package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionBySlugNameQuery struct {
	Slug string `json:"slug"`
}

func (*GetDeviceDefinitionBySlugNameQuery) Key() string { return "GetDeviceDefinitionBySlugQuery" }

type GetDeviceDefinitionBySlugNameQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceDefinitionBySlugNameQueryHandler(cache services.DeviceDefinitionCacheService) GetDeviceDefinitionBySlugNameQueryHandler {
	return GetDeviceDefinitionBySlugNameQueryHandler{
		DDCache: cache,
	}
}

func (ch GetDeviceDefinitionBySlugNameQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionBySlugNameQuery)

	dd, err := ch.DDCache.GetDeviceDefinitionBySlugName(ctx, qry.Slug)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device slug: %s", qry.Slug),
		}
	}

	return dd, nil
}
