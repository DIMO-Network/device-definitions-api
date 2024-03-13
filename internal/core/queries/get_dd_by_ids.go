package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetDeviceDefinitionByIDsQuery struct {
	DeviceDefinitionID []string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionByIDsQuery) Key() string { return "GetDeviceDefinitionByIDsQuery" }

type GetDeviceDefinitionByIDsQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
	log     *zerolog.Logger
}

func NewGetDeviceDefinitionByIDsQueryHandler(cache services.DeviceDefinitionCacheService, log *zerolog.Logger) GetDeviceDefinitionByIDsQueryHandler {
	return GetDeviceDefinitionByIDsQueryHandler{
		DDCache: cache,
		log:     log,
	}
}

func (ch GetDeviceDefinitionByIDsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDsQuery)

	if len(qry.DeviceDefinitionID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("Device Definition Ids is required"),
		}
	}

	response := &grpc.GetDeviceDefinitionResponse{}

	for _, v := range qry.DeviceDefinitionID {
		dd, _ := ch.DDCache.GetDeviceDefinitionByID(ctx, v)

		if dd == nil {
			if len(qry.DeviceDefinitionID) > 1 {
				ch.log.Warn().Str("deviceDefinitionId", v).Msg("Not found - Device Definition")
				continue
			}

			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device definition id: %s", v),
			}
		}

		rp := common.BuildFromQueryResultToGRPC(dd)
		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
