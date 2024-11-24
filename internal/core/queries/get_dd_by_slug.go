package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceDefinitionBySlugQuery struct {
	DefinitionID string `json:"definitionId"`
}

func (*GetDeviceDefinitionBySlugQuery) Key() string { return "GetDeviceDefinitionBySlugQuery" }

type GetDeviceDefinitionBySlugQueryHandler struct {
	definitionsOnChainService gateways.DeviceDefinitionOnChainService
	dbs                       func() *db.ReaderWriter
}

func NewGetDeviceDefinitionBySlugQueryHandler(ddOnChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetDeviceDefinitionBySlugQueryHandler {
	return GetDeviceDefinitionBySlugQueryHandler{
		definitionsOnChainService: ddOnChainSvc,
		dbs:                       dbs,
	}
}

func (ch GetDeviceDefinitionBySlugQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionBySlugQuery)

	dd, err := ch.definitionsOnChainService.GetDefinitionByID(ctx, qry.DefinitionID, ch.dbs().Reader)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find definition id: %s", qry.DefinitionID),
		}
	}

	dbDefinition, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.NameSlug.EQ(dd.ID),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType)).One(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}

	result, err := common.BuildFromDeviceDefinitionToQueryResult(dbDefinition)

	return result, err
}
