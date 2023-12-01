package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
)

type GetAllDeviceDefinitionByMakeYearRangeQuery struct {
	Make      string
	StartYear int32
	EndYear   int32
}

func (*GetAllDeviceDefinitionByMakeYearRangeQuery) Key() string {
	return "GetAllDeviceDefinitionByMakeYearRangeQuery"
}

type GetAllDeviceDefinitionByMakeYearRangeQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewGetAllDeviceDefinitionByMakeYearRangeQueryHandler(repository repositories.DeviceDefinitionRepository) GetAllDeviceDefinitionByMakeYearRangeQueryHandler {
	return GetAllDeviceDefinitionByMakeYearRangeQueryHandler{
		Repository: repository,
	}
}

func (ch GetAllDeviceDefinitionByMakeYearRangeQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetAllDeviceDefinitionByMakeYearRangeQuery)

	all, err := ch.Repository.GetDevicesByMakeYearRange(ctx, qry.Make, qry.StartYear, qry.EndYear)

	if err != nil {
		return nil, err
	}

	response := &grpc.GetDeviceDefinitionResponse{}
	for _, v := range all {
		dd, err := common.BuildFromDeviceDefinitionToQueryResult(v)
		if err != nil {
			return nil, err
		}
		rp := common.BuildFromQueryResultToGRPC(dd)

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
