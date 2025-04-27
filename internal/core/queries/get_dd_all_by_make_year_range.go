package queries

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
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
	dbs        func() *db.ReaderWriter
	onChainSvc gateways.DeviceDefinitionOnChainService
}

func NewGetAllDeviceDefinitionByMakeYearRangeQueryHandler(onChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetAllDeviceDefinitionByMakeYearRangeQueryHandler {
	return GetAllDeviceDefinitionByMakeYearRangeQueryHandler{
		dbs:        dbs,
		onChainSvc: onChainSvc,
	}
}

func (ch GetAllDeviceDefinitionByMakeYearRangeQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetAllDeviceDefinitionByMakeYearRangeQuery)
	manufacturer, err2 := ch.onChainSvc.GetManufacturer(ctx, qry.Make, ch.dbs().Reader)
	if err2 != nil {
		return nil, err2
	}
	// todo need to use queryTableland and do custom query for this one

	all, err := ch.onChainSvc.GetDeviceDefinitions(ctx, manufacturer.TokenID, "", "", qry.StartYear)
	//all, err := ch.Repository.GetDevicesByMakeYearRange(ctx, qry.Make, qry.StartYear, qry.EndYear)

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
