package commands

import (
	"context"

	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/pkg/errors"
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
	dt.Metadatakey = stringutils.SlugString(command.Name)

	err := dt.Insert(ctx, ch.DBS().Writer, boil.Infer())

	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrapf(err, "error inserting device type: %s", command.Name)}
	}

	return CreateDeviceTypeCommandResult{ID: dt.ID}, nil
}
