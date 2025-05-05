//nolint:tagliatelle
package commands

import (
	"context"
	"database/sql"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UpdateDeviceStyleCommand struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ExternalStyleID    string `json:"external_style_id"`
	Source             string `json:"source"`
	SubModel           string `json:"sub_model"`
	HardwareTemplateID string `json:"hardware_template_id,omitempty"`
	DefinitionID       string `json:"definition_id,omitempty"`
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
			ID:           command.ID,
			DefinitionID: command.DefinitionID,
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

	ds.HardwareTemplateID = null.StringFrom(command.HardwareTemplateID)

	if err := ds.Upsert(ctx, ch.DBS().Writer.DB, true, []string{models.DeviceStyleColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return UpdateDeviceStyleCommandResult{ID: ds.ID}, nil
}
