package queries

import (
	"context"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
)

type GetAllDeviceDefinitionOnChainQuery struct {
	MakeSlug           string `json:"makeSlug"`
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	Year               int    `json:"year"`
	Model              string `json:"model"`
	PageIndex          int32  `json:"pageIndex"`
	PageSize           int32  `json:"pageSize"`
}

func (*GetAllDeviceDefinitionOnChainQuery) Key() string { return "GetAllDeviceDefinitionOnChainQuery" }

type GetAllDeviceDefinitionOnChainQueryHandler struct {
	DBS                            func() *db.ReaderWriter
	DeviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService
}

func NewGetAllDeviceDefinitionOnChainQueryHandler(dbs func() *db.ReaderWriter, deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService) GetAllDeviceDefinitionOnChainQueryHandler {
	return GetAllDeviceDefinitionOnChainQueryHandler{
		DBS:                            dbs,
		DeviceDefinitionOnChainService: deviceDefinitionOnChainService,
	}
}

func (ch GetAllDeviceDefinitionOnChainQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetAllDeviceDefinitionOnChainQuery)
	dm, err := ch.DeviceDefinitionOnChainService.GetManufacturer(ctx, qry.MakeSlug, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}
	dmDb, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(qry.MakeSlug)).One(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}

	all, err := ch.DeviceDefinitionOnChainService.GetDeviceDefinitions(ctx, types.NewNullDecimal(decimal.New(int64(dm.TokenID), 0)), qry.DeviceDefinitionID, qry.Model, qry.Year, qry.PageIndex, qry.PageSize)
	if err != nil {
		return nil, err
	}

	response := &grpc.GetDeviceDefinitionResponse{}
	for _, v := range all {
		dd, err := common.BuildFromDeviceDefinitionToQueryResult(&v, &coremodels.DeviceMake{
			ID:              dmDb.ID,
			Name:            dm.Name,
			LogoURL:         dmDb.LogoURL,
			OemPlatformName: dmDb.OemPlatformName,
			NameSlug:        dmDb.NameSlug,
			CreatedAt:       dmDb.CreatedAt,
			UpdatedAt:       dmDb.UpdatedAt,
		}, nil, nil)
		if err != nil {
			return nil, err
		}

		rp := common.BuildFromQueryResultToGRPC(dd)

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}

/*

 */
