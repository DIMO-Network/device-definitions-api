package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type GetReviewsByIDQuery struct {
	ReviewID string `json:"review_id"`
}

func (*GetReviewsByIDQuery) Key() string { return "GetReviewsByIDQuery" }

type GetReviewsByIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetReviewsByIDQueryHandler(dbs func() *db.ReaderWriter) GetReviewsByIDQueryHandler {
	return GetReviewsByIDQueryHandler{DBS: dbs}
}

func (qh GetReviewsByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetReviewsByIDQuery)

	review, err := models.Reviews(models.ReviewWhere.ID.EQ(qry.ReviewID)).
		One(ctx, qh.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find review id: %s", qry.ReviewID),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get review"),
		}
	}

	result := &p_grpc.DeviceReview{
		Id:                 review.ID,
		Url:                review.URL,
		ImageURL:           review.ImageURL,
		Channel:            review.Channel.String,
		DeviceDefinitionId: review.DeviceDefinitionID,
		Comments:           review.Comments,
		Approved:           review.Approved,
		ApprovedBy:         review.ApprovedBy,
	}

	return result, nil
}
