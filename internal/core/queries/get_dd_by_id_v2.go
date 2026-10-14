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
				Err: fmt.Errorf("could not find device definition id: %s: %w", qry.DefinitionID, err),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definition %s from catalog: %w", qry.DefinitionID, err),
		}
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DefinitionID),
		}
	}

	return dd, nil
}
