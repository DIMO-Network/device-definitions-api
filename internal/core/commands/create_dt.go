package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type CreateDeviceTypeCommand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateDeviceTypeCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceTypeCommand) Key() string { return "CreateDeviceTypeCommand" }

type CreateDeviceTypeCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewCreateDeviceTypeCommandHandler(dbs func() *db.ReaderWriter) CreateDeviceTypeCommandHandler {
	return CreateDeviceTypeCommandHandler{DBS: dbs}
}

func (ch CreateDeviceTypeCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateDeviceTypeCommand)

	dt := models.DeviceType{}
	dt.ID = command.ID
	dt.Name = command.Name
	dt.Metadatakey = common.SlugString(command.Name)

	err := dt.Insert(ctx, ch.DBS().Writer, boil.Infer())

	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrapf(err, "error inserting device type: %s", command.Name)}
	}

	return CreateDeviceTypeCommandResult{ID: dt.ID}, nil
}
