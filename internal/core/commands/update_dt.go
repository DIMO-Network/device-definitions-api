package commands

import (
	"context"
	"database/sql"
	"strings"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UpdateDeviceTypeCommand struct {
	ID               string                                  `json:"id"`
	Name             string                                  `json:"name"`
	DeviceAttributes []*coremodels.CreateDeviceTypeAttribute `json:"deviceAttributes"`
}

type UpdateDeviceTypeCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateDeviceTypeCommand) Key() string { return "UpdateDeviceTypeCommand" }

type UpdateDeviceTypeCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewUpdateDeviceTypeCommandHandler(dbs func() *db.ReaderWriter) UpdateDeviceTypeCommandHandler {
	return UpdateDeviceTypeCommandHandler{DBS: dbs}
}

func (ch UpdateDeviceTypeCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceTypeCommand)

	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.ID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if dt == nil {
		dt = &models.DeviceType{
			ID:          command.ID,
			Name:        command.Name,
			Metadatakey: common.SlugString(command.Name),
		}
	}

	dt.Name = command.Name
	// make sure lowercased
	for i, attribute := range command.DeviceAttributes {
		command.DeviceAttributes[i].Name = strings.ToLower(attribute.Name)
	}

	metaData := make(map[string]interface{})
	metaData["properties"] = command.DeviceAttributes

	err = dt.Properties.Marshal(metaData)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	if err := dt.Upsert(ctx, ch.DBS().Writer.DB, true, []string{models.DeviceTypeColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return UpdateDeviceTypeCommandResult{ID: dt.ID}, nil
}
