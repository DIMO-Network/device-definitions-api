//nolint:tagliatelle
package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/big"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UpdateDeviceMakeCommand struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	LogoURL            null.String     `json:"logo_url"`
	OemPlatformName    null.String     `json:"oem_platform_name"`
	TokenID            *big.Int        `json:"tokenId,omitempty"`
	ExternalIDs        json.RawMessage `json:"external_ids"`
	Metadata           json.RawMessage `json:"metadata"`
	HardwareTemplateID string          `json:"hardware_template_id,omitempty"`
}

type UpdateDeviceMakeCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateDeviceMakeCommand) Key() string { return "UpdateDeviceMakeCommand" }

type UpdateDeviceMakeCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewUpdateDeviceMakeCommandHandler(dbs func() *db.ReaderWriter) UpdateDeviceMakeCommandHandler {
	return UpdateDeviceMakeCommandHandler{DBS: dbs}
}

func (ch UpdateDeviceMakeCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceMakeCommand)

	dm, err := models.DeviceMakes(models.DeviceMakeWhere.ID.EQ(command.ID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if dm == nil {
		dm = &models.DeviceMake{
			ID:       command.ID,
			NameSlug: shared.SlugString(command.Name),
		}
	}

	if len(command.Name) > 0 {
		dm.Name = command.Name
	}

	if command.LogoURL.Valid {
		dm.LogoURL = command.LogoURL
	}

	if command.OemPlatformName.Valid {
		dm.OemPlatformName = command.OemPlatformName
	}

	dm.ExternalIds = null.JSONFrom([]byte(command.ExternalIDs))
	dm.Metadata = null.JSONFrom([]byte(command.Metadata))
	dm.HardwareTemplateID = null.StringFrom(command.HardwareTemplateID)

	if err := dm.Upsert(ctx, ch.DBS().Writer.DB, true, []string{models.DeviceMakeColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return UpdateDeviceMakeCommandResult{ID: dm.ID}, nil
}
