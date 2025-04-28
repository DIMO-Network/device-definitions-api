package queries

import (
	"context"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"strings"

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
	makeSlug := shared.SlugString(qry.Make)
	manufacturer, err := ch.onChainSvc.GetManufacturer(ctx, makeSlug, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}

	var conditions []string
	if qry.StartYear > 0 {
		conditions = append(conditions, "year >= "+string(qry.StartYear))
	}
	if qry.EndYear > qry.StartYear {
		conditions = append(conditions, "year <= "+string(qry.EndYear))
	}
	whereClause := strings.Join(conditions, " AND ")
	if whereClause != "" {
		whereClause = " WHERE " + whereClause
	}

	all, err := ch.onChainSvc.QueryDefinitionsCustom(ctx, manufacturer.TokenID, whereClause, 0)
	if err != nil {
		return nil, err
	}
	dmDb, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(makeSlug)).One(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}
	dm := coremodels.ConvertDeviceMakeFromDB(dmDb)

	response := &grpc.GetDeviceDefinitionResponse{}
	for _, v := range all {
		dd, err := common.BuildFromDeviceDefinitionToQueryResult(&v, dm, nil, nil)
		if err != nil {
			return nil, err
		}
		rp := common.BuildFromQueryResultToGRPC(dd)

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
