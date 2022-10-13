package commands

import (
	"context"
	"database/sql"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UpdateDeviceStyleCommand struct {
	ID                 string `json:"id"`
	DeviceDefinitionID string `json:"device_definition_id"`
	Name               string `json:"name"`
	ExternalStyleID    string `json:"external_style_id"`
	Source             string `json:"source"`
	SubModel           string `json:"sub_model"`
}

type UpdateDeviceStyleCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateDeviceStyleCommand) Key() string { return "UpdateDeviceStyleCommand" }

type UpdateDeviceStyleCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewUpdateDeviceStyleCommandHandler(dbs func() *db.ReaderWriter) UpdateDeviceStyleCommandHandler {
	return UpdateDeviceStyleCommandHandler{DBS: dbs}
}

func (ch UpdateDeviceStyleCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceStyleCommand)

	ds, err := models.DeviceStyles(models.DeviceStyleWhere.ID.EQ(command.ID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if ds == nil {
		ds = &models.DeviceStyle{
			ID:                 command.ID,
			DeviceDefinitionID: command.DeviceDefinitionID,
		}
	}

	if len(command.Name) > 0 {
		ds.Name = command.Name
	}

	if len(command.ExternalStyleID) > 0 {
		ds.ExternalStyleID = command.ExternalStyleID
	}

	if len(command.Source) > 0 {
		ds.Source = command.Source
	}

	if len(command.SubModel) > 0 {
		ds.SubModel = command.SubModel
	}

	if err := ds.Upsert(ctx, ch.DBS().Writer.DB, true, []string{models.DeviceStyleColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return UpdateDeviceStyleCommandResult{ID: ds.ID}, nil
}
