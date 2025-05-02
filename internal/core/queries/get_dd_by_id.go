package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionByIDQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
}

func (*GetDeviceDefinitionByIDQuery) Key() string { return "GetDeviceDefinitionByIdQuery" }

type GetDeviceDefinitionByIDQueryHandler struct {
	dbs        func() *db.ReaderWriter
	onChainSvc gateways.DeviceDefinitionOnChainService
}

func NewGetDeviceDefinitionByIDQueryHandler(onChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetDeviceDefinitionByIDQueryHandler {
	return GetDeviceDefinitionByIDQueryHandler{
		onChainSvc: onChainSvc,
		dbs:        dbs,
	}
}

func (ch GetDeviceDefinitionByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDQuery)

	dd, _, err := ch.onChainSvc.GetDefinitionByID(ctx, qry.DeviceDefinitionID, ch.dbs().Reader)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	return dd, nil
}
