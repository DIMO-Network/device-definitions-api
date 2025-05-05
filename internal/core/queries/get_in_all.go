package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetAllIntegrationQuery struct {
}

func (*GetAllIntegrationQuery) Key() string { return "GetAllIntegrationQuery" }

type GetAllIntegrationQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetAllIntegrationQueryHandler(dbs func() *db.ReaderWriter) GetAllIntegrationQueryHandler {
	return GetAllIntegrationQueryHandler{DBS: dbs}
}

func (ch GetAllIntegrationQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {
	all, err := models.Integrations(qm.Limit(100)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integrations"),
		}
	}

	result := make([]coremodels.GetIntegrationQueryResult, len(all))
	for i, v := range all {
		im := new(coremodels.IntegrationsMetadata)
		if v.Metadata.Valid {
			err = v.Metadata.Unmarshal(&im)

			if err != nil {
				return nil, &exceptions.InternalError{
					Err: fmt.Errorf("failed to unmarshall integration metadata id %s", v.ID),
				}
			}
		}

		result[i] = coremodels.GetIntegrationQueryResult{
			ID:                      v.ID,
			Type:                    v.Type,
			Style:                   v.Style,
			Vendor:                  v.Vendor,
			AutoPiDefaultTemplateID: im.AutoPiDefaultTemplateID,
			RefreshLimitSecs:        v.RefreshLimitSecs,
			TokenID:                 v.TokenID.Int,
			Points:                  v.Points,
		}

		if !v.ManufacturerTokenID.IsZero() {
			result[i].ManufacturerTokenID = v.ManufacturerTokenID.Int
		}

		if im.AutoPiPowertrainToTemplateID != nil {
			result[i].AutoPiPowertrainToTemplateID = im.AutoPiPowertrainToTemplateID
		}
	}

	return result, nil
}
