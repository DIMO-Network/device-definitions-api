package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceDefinitionWithoutImageQuery struct {
}

func (*GetDeviceDefinitionWithoutImageQuery) Key() string {
	return "GetDeviceDefinitionWithoutImageQuery"
}

type GetDeviceDefinitionWithoutImageQueryHandler struct {
	DBS func() *db.ReaderWriter
	log *zerolog.Logger
}

func NewGetDeviceDefinitionWithoutImageQueryHandler(dbs func() *db.ReaderWriter, log *zerolog.Logger) GetDeviceDefinitionWithoutImageQueryHandler {
	return GetDeviceDefinitionWithoutImageQueryHandler{
		DBS: dbs,
		log: log,
	}
}

func (ch GetDeviceDefinitionWithoutImageQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, err := repoModel.DeviceDefinitions(qm.Load(repoModel.DeviceDefinitionRels.DeviceMake),
		repoModel.DeviceDefinitionWhere.ImageURL.IsNull()).All(ctx, ch.DBS().Reader)

	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definitions"),
		}
	}

	response := &grpc.GetDeviceDefinitionResponse{}

	for _, v := range all {
		ch.log.Info().Msg(fmt.Sprintf("Start %s", v.ID))

		dd := common.BuildFromDeviceDefinitionToQueryResult(v)
		ch.log.Info().Msg(fmt.Sprintf("DD %s", dd.DeviceDefinitionID))
		rp := common.BuildFromQueryResultToGRPC(dd)
		ch.log.Info().Msg(fmt.Sprintf("GRPC %s", rp.DeviceDefinitionId))

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
