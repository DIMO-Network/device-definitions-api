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
	ddCatalogSvc gateways.DeviceDefinitionCatalogService
	dbs          func() *db.ReaderWriter
}

func NewGetDeviceDefinitionByIDQueryV2Handler(ddCatalogSvc gateways.DeviceDefinitionCatalogService, dbs func() *db.ReaderWriter) GetDeviceDefinitionByIDQueryV2Handler {
	return GetDeviceDefinitionByIDQueryV2Handler{
		ddCatalogSvc: ddCatalogSvc,
		dbs:          dbs,
	}
}

func (ch GetDeviceDefinitionByIDQueryV2Handler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDQueryV2)

	dd, _, err := ch.ddCatalogSvc.GetTemplateByID(ctx, qry.DefinitionID)

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
