package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type DeleteIntegrationFeatureCommand struct {
	ID string `json:"id"`
}

type DeleteIntegrationFeatureResult struct {
	ID string `json:"id"`
}

func (*DeleteIntegrationFeatureCommand) Key() string { return "DeleteIntegrationFeatureCommand" }

type DeleteIntegrationFeatureCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewDeleteIntegrationFeatureCommandHandler(dbs func() *db.ReaderWriter) DeleteIntegrationFeatureCommandHandler {
	return DeleteIntegrationFeatureCommandHandler{DBS: dbs}
}

func (ch DeleteIntegrationFeatureCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*DeleteIntegrationFeatureCommand)

	feature, err := models.IntegrationFeatures(models.IntegrationFeatureWhere.FeatureKey.EQ(command.ID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find integration feature id: %s", command.ID),
			}
		}
	}

	if _, err := feature.Delete(ctx, ch.DBS().Writer.DB); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return DeleteIntegrationFeatureCommand{ID: feature.FeatureKey}, nil
}
