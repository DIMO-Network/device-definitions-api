package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionBySlugQuery struct {
	// Slug is the model slug
	Slug string `json:"slug"`
	Year int    `json:"year"`
}

func (*GetDeviceDefinitionBySlugQuery) Key() string { return "GetDeviceDefinitionBySlugQuery" }

type GetDeviceDefinitionBySlugQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceDefinitionBySlugQueryHandler(cache services.DeviceDefinitionCacheService) GetDeviceDefinitionBySlugQueryHandler {
	return GetDeviceDefinitionBySlugQueryHandler{
		DDCache: cache,
	}
}

func (ch GetDeviceDefinitionBySlugQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionBySlugQuery)

	dd, err := ch.DDCache.GetDeviceDefinitionBySlug(ctx, qry.Slug, qry.Year)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device slug: %s and year: %d", qry.Slug, qry.Year),
		}
	}

	return dd, nil
}
