package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceDefinitionBySourceQuery struct {
	Source string `json:"source" validate:"required"`
}

func (*GetDeviceDefinitionBySourceQuery) Key() string {
	return "GetDeviceDefinitionBySourceQuery"
}

type GetDeviceDefinitionBySourceQueryHandler struct {
	DBS func() *db.ReaderWriter
	log *zerolog.Logger
}

func NewGetDeviceDefinitionBySourceQueryHandler(dbs func() *db.ReaderWriter, log *zerolog.Logger) GetDeviceDefinitionBySourceQueryHandler {
	return GetDeviceDefinitionBySourceQueryHandler{
		DBS: dbs,
		log: log,
	}
}

func (ch GetDeviceDefinitionBySourceQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionBySourceQuery)

	all, err := repoModel.DeviceDefinitions(
		qm.Where("external_ids->>? IS NOT NULL", qry.Source),
		qm.Load(repoModel.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(repoModel.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(repoModel.DeviceDefinitionRels.DeviceIntegrations, repoModel.DeviceIntegrationRels.Integration)),
		qm.Load(repoModel.DeviceDefinitionRels.DeviceStyles),
		qm.Load(repoModel.DeviceDefinitionRels.DeviceType)).All(ctx, ch.DBS().Reader)

	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definitions"),
		}
	}

	response := &grpc.GetDeviceDefinitionResponse{}

	for _, v := range all {
		dd, err := common.BuildFromDeviceDefinitionToQueryResult(v)
		if err != nil {
			return nil, err
		}
		rp := common.BuildFromQueryResultToGRPC(dd)
		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
