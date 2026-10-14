package queries

import (
	"context"
	"errors"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionByIDQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
}

func (*GetDeviceDefinitionByIDQuery) Key() string { return "GetDeviceDefinitionByIdQuery" }

type GetDeviceDefinitionByIDQueryHandler struct {
	dbs        func() *db.ReaderWriter
	catalogSvc gateways.DeviceDefinitionCatalogService
}

func NewGetDeviceDefinitionByIDQueryHandler(catalogSvc gateways.DeviceDefinitionCatalogService, dbs func() *db.ReaderWriter) GetDeviceDefinitionByIDQueryHandler {
	return GetDeviceDefinitionByIDQueryHandler{
		catalogSvc: catalogSvc,
		dbs:        dbs,
	}
}

func (ch GetDeviceDefinitionByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIDQuery)

	dd, _, err := ch.catalogSvc.GetTemplateByID(ctx, qry.DeviceDefinitionID)

	if err != nil {
		// GetTemplateByID reports a genuine "does not exist" as the wrapped
		// ErrTemplateNotFound sentinel, and the HTTP layer maps a not-found by
		// type assertion on *exceptions.NotFoundError, which a wrapped
		// sentinel never satisfies. Returning it raw answered 500 for every
		// unknown id, where the pre-migration on-chain path answered 404.
		// Anything else is a catalog outage and stays an internal error: a
		// caller reacting to a spurious 404 could create a duplicate
		// definition for a vehicle that already exists.
		if errors.Is(err, gateways.ErrTemplateNotFound) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device definition id: %s: %w", qry.DeviceDefinitionID, err),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definition %s from catalog: %w", qry.DeviceDefinitionID, err),
		}
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	return dd, nil
}
