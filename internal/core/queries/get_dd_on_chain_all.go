package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/ericlagergren/decimal"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
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
	ddCache                        services.DeviceDefinitionCacheService
}

func NewGetAllDeviceDefinitionOnChainQueryHandler(dbs func() *db.ReaderWriter, deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService,
	ddCache services.DeviceDefinitionCacheService) GetAllDeviceDefinitionOnChainQueryHandler {
	return GetAllDeviceDefinitionOnChainQueryHandler{
		DBS:                            dbs,
		DeviceDefinitionOnChainService: deviceDefinitionOnChainService,
		ddCache:                        ddCache,
	}
}

func (ch GetAllDeviceDefinitionOnChainQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetAllDeviceDefinitionOnChainQuery)

	dm, err := ch.ddCache.GetDeviceMakeByName(ctx, qry.MakeSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make slug: %s", qry.MakeSlug),
			}
		}
		return nil, err
	}

	all, err := ch.DeviceDefinitionOnChainService.GetDeviceDefinitions(ctx, types.NewNullDecimal(decimal.New(dm.TokenID.Int64(), 0)), qry.DeviceDefinitionID, qry.Model, qry.Year, qry.PageIndex, qry.PageSize)
	if err != nil {
		return nil, err
	}

	response := &grpc.GetDeviceDefinitionResponse{}
	for _, v := range all {

		v.R = v.R.NewStruct()
		v.R.DeviceMake = &models.DeviceMake{
			ID:              dm.ID,
			Name:            dm.Name,
			CreatedAt:       dm.CreatedAt,
			UpdatedAt:       dm.UpdatedAt,
			LogoURL:         dm.LogoURL,
			OemPlatformName: dm.OemPlatformName,
			NameSlug:        dm.NameSlug,
			Metadata:        null.JSONFrom(dm.Metadata),
		}
		v.R.DeviceType = &models.DeviceType{
			Metadatakey: common.VehicleMetadataKey,
		}
		dd, err := common.BuildFromDeviceDefinitionToQueryResult(v)
		if err != nil {
			return nil, err
		}
		rp := common.BuildFromQueryResultToGRPC(dd)

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
