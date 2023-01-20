package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/segmentio/ksuid"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
	// todo need to rethink this method, as this should now update a specific image or insert new images in the images table
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
	img := models.Image{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: command.DeviceDefinitionID,
		FuelAPIID:          null.String{},
		Width:              null.Int{},
		Height:             null.Int{},
		SourceURL:          command.ImageURL,
		DimoS3URL:          null.String{},
		Color:              "default",
		NotExactImage:      false,
	}

	err = img.Upsert(ctx, ch.DBS().Writer.DB, false, []string{models.ImageColumns.DeviceDefinitionID, models.ImageColumns.SourceURL},
		boil.Infer(), boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, command.DeviceDefinitionID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year))
	ch.DDCache.DeleteDeviceDefinitionCacheBySlug(ctx, dd.R.DeviceMake.NameSlug, int(dd.Year))

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
