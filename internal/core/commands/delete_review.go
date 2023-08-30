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
)

type DeleteReviewCommand struct {
	ReviewID string `json:"review_id"`
}

type DeleteReviewCommandResult struct {
	ID string `json:"id"`
}

func (*DeleteReviewCommand) Key() string { return "DeleteReviewCommand" }

type DeleteReviewCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewDeleteReviewCommandHandler(dbs func() *db.ReaderWriter) DeleteReviewCommandHandler {
	return DeleteReviewCommandHandler{DBS: dbs}
}

func (ch DeleteReviewCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*DeleteReviewCommand)

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

	if _, err := review.Delete(ctx, ch.DBS().Writer.DB); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return DeleteReviewCommandResult{ID: review.ID}, nil
}
