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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UpdateReviewCommand struct {
	ReviewID string `json:"review_id"`
	URL      string `json:"url"`
	ImageURL string `json:"imageURL"`
	Channel  string `json:"channel"`
	Comments string `json:"comments"`
}

type UpdateReviewCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateReviewCommand) Key() string { return "UpdateReviewCommand" }

type UpdateReviewCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewUpdateReviewCommandHandler(dbs func() *db.ReaderWriter) UpdateReviewCommandHandler {
	return UpdateReviewCommandHandler{DBS: dbs}
}

func (ch UpdateReviewCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateReviewCommand)

	review, err := models.Reviews(models.ReviewWhere.ID.EQ(command.ReviewID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find review id: %s", command.ReviewID),
			}
		}
	}

	review.ImageURL = command.ImageURL
	review.URL = command.URL
	review.Channel = null.StringFrom(command.Channel)
	review.Comments = command.Comments

	if _, err := review.Update(ctx, ch.DBS().Writer.DB, boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return UpdateReviewCommandResult{ID: review.ID}, nil
}
