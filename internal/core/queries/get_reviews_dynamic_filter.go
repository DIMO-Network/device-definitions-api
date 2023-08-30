package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetReviewsDynamicFilterQuery struct {
	MakeID             string `json:"make_id"`
	DeviceDefinitionID string `json:"device_definition_id"`
	Year               int    `json:"year"`
	Model              string `json:"model"`
	Approved           bool   `json:"approved"`
	PageIndex          int    `json:"page_index"`
	PageSize           int    `json:"page_size"`
}

func (*GetReviewsDynamicFilterQuery) Key() string {
	return "GetReviewsDynamicFilterQuery"
}

type GetReviewsDynamicFilterQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetReviewsDynamicFilterQueryHandler(dbs func() *db.ReaderWriter) GetReviewsDynamicFilterQueryHandler {
	return GetReviewsDynamicFilterQueryHandler{DBS: dbs}
}

func (qh GetReviewsDynamicFilterQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetReviewsDynamicFilterQuery)

	var queryMods []qm.QueryMod

	queryMods = append(queryMods, models.ReviewWhere.Approved.EQ(qry.Approved))

	if len(qry.DeviceDefinitionID) > 1 {
		queryMods = append(queryMods, models.ReviewWhere.DeviceDefinitionID.EQ(string(qry.DeviceDefinitionID)))
	}

	if len(qry.MakeID) > 1 || qry.Year > 1980 && qry.Year < 2999 || len(qry.Model) > 1 {
		queryJoin := qm.InnerJoin("device_definitions_api.device_definitions dd on dd.id = reviews.device_definition_id")
		queryMods = append(queryMods, queryJoin)

		if len(qry.MakeID) > 1 {
			queryMods = append(queryMods, qm.And("dd.device_make_id = ?", qry.MakeID))
		}

		if qry.Year > 1980 && qry.Year < 2999 {
			queryMods = append(queryMods, qm.And("dd.year = ?", qry.Year))
		}

		if len(qry.Model) > 1 {
			queryMods = append(queryMods, qm.And("dd.model = ?", qry.Model))
		}
	}

	queryMods = append(queryMods,
		qm.Load(models.ReviewRels.DeviceDefinition),
		qm.Load(qm.Rels(models.ReviewRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake)),
		qm.Limit(qry.PageSize),
		qm.Offset(qry.PageIndex*qry.PageSize))

	all, err := models.Reviews(queryMods...).All(ctx, qh.DBS().Reader)

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
			Name:               common.BuildDeviceDefinitionName(review.R.DeviceDefinition.Year, review.R.DeviceDefinition.R.DeviceMake.Name, review.R.DeviceDefinition.Model),
		})
	}

	return result, nil
}
