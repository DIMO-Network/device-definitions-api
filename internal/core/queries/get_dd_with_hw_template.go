package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
)

type GetDefinitionsWithHWTemplateQuery struct {
}

func (*GetDefinitionsWithHWTemplateQuery) Key() string {
	return "GetDefinitionsWithHWTemplateQuery"
}

type GetDefinitionsWithHWTemplateQueryHandler struct {
	dbs func() *db.ReaderWriter
	log *zerolog.Logger
}

func NewGetDefinitionsWithHWTemplateQueryHandler(dbs func() *db.ReaderWriter, log *zerolog.Logger) GetDefinitionsWithHWTemplateQueryHandler {
	return GetDefinitionsWithHWTemplateQueryHandler{
		dbs: dbs,
		log: log,
	}
}

func (ch GetDefinitionsWithHWTemplateQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {

	response := &grpc.GetDevicesMMYResponse{}

	all, err := models.DeviceDefinitions(
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		models.DeviceDefinitionWhere.HardwareTemplateID.IsNotNull()).All(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}
	response.Device = make([]*grpc.GetDevicesMMYItemResponse, len(all))
	for i, definition := range all {
		response.Device[i] = &grpc.GetDevicesMMYItemResponse{
			Make:               definition.R.DeviceMake.Name,
			Model:              definition.Model,
			Year:               int32(definition.Year),
			Id:                 definition.ID,
			HardwareTemplateId: definition.HardwareTemplateID.String,
		}
	}

	return response, nil
}
