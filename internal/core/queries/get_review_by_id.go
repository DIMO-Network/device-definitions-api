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

type GetReviewsByModelQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionID"`
}

func (*GetReviewsByModelQuery) Key() string { return "GetReviewsByModelQuery" }

type GetReviewsByModelQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetReviewsByModelQueryHandler(dbs func() *db.ReaderWriter) GetReviewsByModelQueryHandler {
	return GetReviewsByModelQueryHandler{DBS: dbs}
}

func (qh GetReviewsByModelQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetReviewsByModelQuery)

	all, err := models.Reviews(models.ReviewWhere.DeviceDefinitionID.EQ(qry.DeviceDefinitionID)).
		All(ctx, qh.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &p_grpc.GetReviewsResponse{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get reviews"),
		}
	}

	result := &p_grpc.GetReviewsResponse{}

	for _, v := range all {
		result.Reviews = append(result.Reviews, &p_grpc.DeviceReview{
			Url:                v.URL,
			ImageURL:           v.ImageURL,
			Channel:            v.Channel.String,
			DeviceDefinitionId: v.DeviceDefinitionID,
		})
	}

	return result, nil
}
