package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceDefinitionByMakeModelYearQuery struct {
	Make  string `json:"make" validate:"required"`
	Model string `json:"model" validate:"required"`
	Year  int    `json:"year" validate:"required"`
}

func (*GetDeviceDefinitionByMakeModelYearQuery) Key() string {
	return "GetDeviceDefinitionByMakeModelYearQuery"
}

type GetDeviceDefinitionByMakeModelYearQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceDefinitionByMakeModelYearQueryHandler(cache services.DeviceDefinitionCacheService) GetDeviceDefinitionByMakeModelYearQueryHandler {
	return GetDeviceDefinitionByMakeModelYearQueryHandler{
		DDCache: cache,
	}
}

func (ch GetDeviceDefinitionByMakeModelYearQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByMakeModelYearQuery)

	dd, _ := ch.DDCache.GetDeviceDefinitionByMakeModelAndYears(ctx, qry.Make, qry.Model, qry.Year)

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition with MMY: %s %s %d", qry.Make, qry.Model, qry.Year),
		}
	}

	return dd, nil
}
