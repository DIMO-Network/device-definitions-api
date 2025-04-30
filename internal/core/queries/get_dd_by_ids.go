package queries

import (
	"context"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"

	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetDeviceDefinitionByIDsQuery struct {
	DeviceDefinitionID []string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionByIDsQuery) Key() string { return "GetDeviceDefinitionByIDsQuery" }

type GetDeviceDefinitionByIDsQueryHandler struct {
	dbs        func() *db.ReaderWriter
	log        *zerolog.Logger
	onChainSvc gateways.DeviceDefinitionOnChainService
}

func NewGetDeviceDefinitionByIDsQueryHandler(log *zerolog.Logger, onChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetDeviceDefinitionByIDsQueryHandler {
	return GetDeviceDefinitionByIDsQueryHandler{
		onChainSvc: onChainSvc,
		log:        log,
		dbs:        dbs,
	}
}

// Handle gets device definition based on legacy KSUID id
func (ch GetDeviceDefinitionByIDsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	// desired response type GetDeviceDefinitionResponse
	qry := query.(*GetDeviceDefinitionByIDsQuery)

	if len(qry.DeviceDefinitionID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("Device Definition Ids is required"),
		}
	}
	response := &p_grpc.GetDeviceDefinitionResponse{
		DeviceDefinitions: make([]*p_grpc.GetDeviceDefinitionItemResponse, 0),
	}

	for _, ddid := range qry.DeviceDefinitionID {
		dd, manufID, err := ch.onChainSvc.GetDefinitionByID(ctx, ddid, ch.dbs().Reader)
		if err != nil {
			return nil, err
		}
		if dd == nil {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device definition id: %s", ddid),
			}
		}
		// todo refactor this out, same pattern in a couple places
		makeName, err := ch.onChainSvc.GetManufacturerNameByID(ctx, manufID)
		if err != nil {
			return nil, err
		}
		dm, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(makeName)).One(ctx, ch.dbs().Reader)
		if err != nil {
			return nil, err
		}
		dss, _ := models.DeviceStyles(models.DeviceStyleWhere.DefinitionID.EQ(ddid)).All(ctx, ch.dbs().Reader)
		trx, _ := models.DefinitionTransactions(models.DefinitionTransactionWhere.DefinitionID.EQ(ddid)).All(ctx, ch.dbs().Reader)
		rp, err := common.BuildFromDeviceDefinitionToQueryResult(dd, coremodels.ConvertDeviceMakeFromDB(dm), dss, trx)
		if err != nil {
			return nil, err
		}
		gg := common.BuildFromQueryResultToGRPC(rp)
		response.DeviceDefinitions = append(response.DeviceDefinitions, gg)
	}

	return response, nil
}
