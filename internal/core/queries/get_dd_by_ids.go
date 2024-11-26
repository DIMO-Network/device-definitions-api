package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetDeviceDefinitionByIDsQuery struct {
	DeviceDefinitionID []string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionByIDsQuery) Key() string { return "GetDeviceDefinitionByIDsQuery" }

type GetDeviceDefinitionByIDsQueryHandler struct {
	DDCache    services.DeviceDefinitionCacheService
	log        *zerolog.Logger
	repository repositories.DeviceDefinitionRepository
}

func NewGetDeviceDefinitionByIDsQueryHandler(cache services.DeviceDefinitionCacheService, log *zerolog.Logger,
	repository repositories.DeviceDefinitionRepository) GetDeviceDefinitionByIDsQueryHandler {
	return GetDeviceDefinitionByIDsQueryHandler{
		DDCache:    cache,
		log:        log,
		repository: repository,
	}
}

// Handle gets device definition based on legacy KSUID id
func (ch GetDeviceDefinitionByIDsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDsQuery)

	if len(qry.DeviceDefinitionID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("Device Definition Ids is required"),
		}
	}

	dd, err := ch.repository.GetByID(ctx, qry.DeviceDefinitionID[0])
	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID[0]),
		}
	}

	rp, err := common.BuildFromDeviceDefinitionToQueryResult(dd)
	if err != nil {
		return nil, err
	}

	return common.BuildFromQueryResultToGRPC(rp), nil
}
