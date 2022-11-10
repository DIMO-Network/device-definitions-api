package queries

import (
	"context"
	"database/sql"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type GetIntegrationFeatureByIDQuery struct {
	ID string `json:"id"`
}

func (*GetIntegrationFeatureByIDQuery) Key() string { return "GetIntegrationFeatureByIDQuery" }

type GetIntegrationFeatureByIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetIntegrationFeatureByIDQueryHandler(dbs func() *db.ReaderWriter) GetIntegrationFeatureByIDQueryHandler {
	return GetIntegrationFeatureByIDQueryHandler{DBS: dbs}
}

func (ch GetIntegrationFeatureByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetIntegrationFeatureByIDQuery)

	feature, err := models.IntegrationFeatures(models.IntegrationFeatureWhere.FeatureKey.EQ(qry.ID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find integration feature id: %s", qry.ID),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration feature"),
		}
	}

	result := coremodels.GetIntegrationFeatureQueryResult{
		FeatureKey:      feature.FeatureKey,
		ElasticProperty: feature.ElasticProperty,
		DisplayName:     feature.DisplayName,
		CreatedAt:       feature.CreatedAt,
		UpdatedAt:       feature.UpdatedAt,
	}

	if feature.CSSIcon.Valid {
		result.CSSIcon = feature.CSSIcon.String
	}

	if feature.FeatureWeight.Valid {
		result.FeatureWeight = feature.FeatureWeight.Float64
	}

	return result, nil
}
