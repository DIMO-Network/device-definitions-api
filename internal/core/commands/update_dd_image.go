package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UpdateDeviceDefinitionImageCommand struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	ImageURL           string `json:"image_url"`
}

type UpdateDeviceDefinitionImageCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateDeviceDefinitionImageCommand) Key() string { return "UpdateDeviceDefinitionImageCommand" }

type UpdateDeviceDefinitionImageCommandHandler struct {
	DBS     func() *db.ReaderWriter
	DDCache services.DeviceDefinitionCacheService
}

func NewUpdateDeviceDefinitionImageCommandHandler(dbs func() *db.ReaderWriter, cache services.DeviceDefinitionCacheService) UpdateDeviceDefinitionImageCommandHandler {
	return UpdateDeviceDefinitionImageCommandHandler{DBS: dbs, DDCache: cache}
}

func (ch UpdateDeviceDefinitionImageCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceDefinitionImageCommand)

	dd, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.ID.EQ(command.DeviceDefinitionID),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	).One(ctx, ch.DBS().Writer)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if err != nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", command.DeviceDefinitionID),
		}
	}

	dd.ImageURL = null.StringFrom(command.ImageURL)

	_, err = dd.Update(ctx, ch.DBS().Writer.DB, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, command.DeviceDefinitionID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year))

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
