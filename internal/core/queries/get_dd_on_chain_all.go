package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
)

type GetAllDeviceDefinitionOnChainQuery struct {
	MakeSlug           string `json:"make_slug"`
	DeviceDefinitionID string `json:"device_definition_id"`
	Year               int    `json:"year"`
	Model              string `json:"model"`
	PageIndex          int32  `json:"page_index"`
	PageSize           int32  `json:"page_size"`
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

	make, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(qry.MakeSlug)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make slug: %s", qry.MakeSlug),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}

	all, err := ch.DeviceDefinitionOnChainService.GetDeviceDefinitions(make.TokenID, qry.DeviceDefinitionID, qry.Model, qry.Year, qry.PageIndex, qry.PageSize)
	if err != nil {
		return nil, err
	}

	response := &grpc.GetDeviceDefinitionResponse{}
	for _, v := range all {

		v.R = v.R.NewStruct()
		v.R.DeviceMake = make

		dd, err := common.BuildFromDeviceDefinitionOnChainToQueryResult(v)
		if err != nil {
			return nil, err
		}
		rp := common.BuildFromQueryResultToGRPC(dd)

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
