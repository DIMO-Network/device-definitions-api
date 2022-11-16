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
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ApproveReviewCommand struct {
	ReviewID   string `json:"review_id"`
	ApprovedBy string `json:"approved_by"`
}

type ApproveReviewCommandResult struct {
	ID string `json:"id"`
}

func (*ApproveReviewCommand) Key() string { return "ApproveReviewCommand" }

type ApproveReviewCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewApproveReviewCommandHandler(dbs func() *db.ReaderWriter) ApproveReviewCommandHandler {
	return ApproveReviewCommandHandler{DBS: dbs}
}

func (ch ApproveReviewCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*ApproveReviewCommand)

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

	review.Approved = true
	review.ApprovedBy = command.ApprovedBy

	if _, err := review.Update(ctx, ch.DBS().Writer.DB, boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return ApproveReviewCommandResult{ID: review.ID}, nil
}
