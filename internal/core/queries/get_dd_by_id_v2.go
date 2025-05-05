package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionByIDQueryV2 struct {
	DefinitionID string `json:"definitionId"`
}

func (*GetDeviceDefinitionByIDQueryV2) Key() string { return "GetDeviceDefinitionByIdQueryV2" }

type GetDeviceDefinitionByIDQueryV2Handler struct {
	ddOnChainSvc gateways.DeviceDefinitionOnChainService
	dbs          func() *db.ReaderWriter
}

func NewGetDeviceDefinitionByIDQueryV2Handler(ddOnChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetDeviceDefinitionByIDQueryV2Handler {
	return GetDeviceDefinitionByIDQueryV2Handler{
		ddOnChainSvc: ddOnChainSvc,
		dbs:          dbs,
	}
}

func (ch GetDeviceDefinitionByIDQueryV2Handler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDQueryV2)

	dd, _, err := ch.ddOnChainSvc.GetDefinitionByID(ctx, qry.DefinitionID, ch.dbs().Reader)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DefinitionID),
		}
	}

	return dd, nil
}
