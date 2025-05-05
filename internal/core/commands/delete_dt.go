package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
)

type DeleteDeviceTypeCommand struct {
	ID string `json:"id"`
}

type DeleteDeviceTypeCommandResult struct {
	ID string `json:"id"`
}

func (*DeleteDeviceTypeCommand) Key() string { return "DeleteDeviceTypeCommand" }

type DeleteDeviceTypeCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewDeleteDeviceTypeCommandHandler(dbs func() *db.ReaderWriter) DeleteDeviceTypeCommandHandler {
	return DeleteDeviceTypeCommandHandler{DBS: dbs}
}

func (ch DeleteDeviceTypeCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*DeleteDeviceTypeCommand)

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.ID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if err != nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device type id: %s", command.ID),
		}
	}

	_, err = dt.Delete(ctx, ch.DBS().Writer.DB)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return DeleteDeviceTypeCommandResult{ID: dt.ID}, nil
}
