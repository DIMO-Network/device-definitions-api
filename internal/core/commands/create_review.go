//nolint:tagliatelle
package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type CreateReviewCommand struct {
	DeviceDefinitionID string `json:"device_definition_id"`
	URL                string `json:"url"`
	ImageURL           string `json:"imageURL"`
	Channel            string `json:"channel"`
	Comments           string `json:"comments"`
}

type CreateReviewCommandResult struct {
	ID string `json:"id"`
}

func (*CreateReviewCommand) Key() string { return "CreateReviewCommand" }

type CreateReviewCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewCreateReviewCommandHandler(dbs func() *db.ReaderWriter) CreateReviewCommandHandler {
	return CreateReviewCommandHandler{DBS: dbs}
}

func (ch CreateReviewCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateReviewCommand)

	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.EQ(command.DeviceDefinitionID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device definition id: %s", command.DeviceDefinitionID),
			}
		}
	}

	review := models.Review{}
	review.ID = ksuid.New().String()
	review.DeviceDefinitionID = dd.ID
	review.URL = command.URL
	review.ImageURL = command.ImageURL
	review.Channel = null.StringFrom(command.Channel)
	review.Comments = command.Comments

	err = review.Insert(ctx, ch.DBS().Writer, boil.Infer())

	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrapf(err, "error inserting review: %s", review.DeviceDefinitionID)}
	}

	return CreateReviewCommandResult{ID: review.ID}, nil
}
