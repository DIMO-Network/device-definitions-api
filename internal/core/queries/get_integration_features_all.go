package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
)

type GetAllIntegrationFeatureQuery struct {
}

func (*GetAllIntegrationFeatureQuery) Key() string { return "GetAllIntegrationFeatureQuery" }

type GetAllIntegrationFeatureQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetAllIntegrationFeatureQuery(dbs func() *db.ReaderWriter) GetAllIntegrationFeatureQueryHandler {
	return GetAllIntegrationFeatureQueryHandler{DBS: dbs}
}

func (ch GetAllIntegrationFeatureQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {

	all, err := models.IntegrationFeatures().All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration features"),
		}
	}

	result := make([]coremodels.GetIntegrationFeatureQueryResult, len(all))
	for i, v := range all {
		result[i] = coremodels.GetIntegrationFeatureQueryResult{
			FeatureKey:      v.FeatureKey,
			ElasticProperty: v.ElasticProperty,
			DisplayName:     v.DisplayName,
			CreatedAt:       v.CreatedAt,
			UpdatedAt:       v.UpdatedAt,
		}

		if v.CSSIcon.Valid {
			result[i].CSSIcon = v.CSSIcon.String
		}

		if v.FeatureWeight.Valid {
			result[i].FeatureWeight = v.FeatureWeight.Float64
		}
	}

	return result, nil
}
