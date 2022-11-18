package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type GetReviewsByDeviceDefinitionIDQuery struct {
	DeviceDefinitionID string `json:"device_definition_id"`
}

func (*GetReviewsByDeviceDefinitionIDQuery) Key() string {
	return "GetReviewsByDeviceDefinitionIDQuery"
}

type GetReviewsByDeviceDefinitionIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetReviewsByDeviceDefinitionIDQueryHandler(dbs func() *db.ReaderWriter) GetReviewsByDeviceDefinitionIDQueryHandler {
	return GetReviewsByDeviceDefinitionIDQueryHandler{DBS: dbs}
}

func (qh GetReviewsByDeviceDefinitionIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetReviewsByDeviceDefinitionIDQuery)

	all, err := models.Reviews(models.ReviewWhere.DeviceDefinitionID.EQ(qry.DeviceDefinitionID),
		qm.Load(models.ReviewRels.DeviceDefinition),
		qm.Load(qm.Rels(models.ReviewRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake))).
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

	for _, review := range all {
		result.Reviews = append(result.Reviews, &p_grpc.DeviceReview{
			Id:                 review.ID,
			Url:                review.URL,
			ImageURL:           review.ImageURL,
			Channel:            review.Channel.String,
			DeviceDefinitionId: review.DeviceDefinitionID,
			Comments:           review.Comments,
			Approved:           review.Approved,
			ApprovedBy:         review.ApprovedBy,
			Name:               fmt.Sprintf("%d %s %s", review.R.DeviceDefinition.Year, review.R.DeviceDefinition.R.DeviceMake.Name, review.R.DeviceDefinition.Model),
		})
	}

	return result, nil
}
