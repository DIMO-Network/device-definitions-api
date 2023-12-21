package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
)

type GetIntegrationByIDQuery struct {
	IntegrationID []string `json:"integration_id" validate:"required"`
}

func (*GetIntegrationByIDQuery) Key() string { return "GetIntegrationByIDQuery" }

type GetIntegrationByIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetIntegrationByIDQueryHandler(dbs func() *db.ReaderWriter) GetIntegrationByIDQueryHandler {
	return GetIntegrationByIDQueryHandler{DBS: dbs}
}

func (ch GetIntegrationByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetIntegrationByIDQuery)

	if len(qry.IntegrationID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("IntegrationID is required"),
		}
	}

	v, err := models.Integrations(models.IntegrationWhere.ID.EQ(qry.IntegrationID[0])).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find integration id: %s", qry.IntegrationID),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration"),
		}
	}

	result := coremodels.GetIntegrationQueryResult{}
	im := new(coremodels.IntegrationsMetadata)
	if v.Metadata.Valid {
		err = v.Metadata.Unmarshal(&im)

		if err != nil {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to unmarshall integration metadata id %s", v.ID),
			}
		}
	}

	result = coremodels.GetIntegrationQueryResult{
		ID:                      v.ID,
		Type:                    v.Type,
		Style:                   v.Style,
		Vendor:                  v.Vendor,
		AutoPiDefaultTemplateID: im.AutoPiDefaultTemplateID,
		RefreshLimitSecs:        v.RefreshLimitSecs,
		TokenID:                 v.TokenID.Int,
		//Points:                  v.Points,
	}

	//if !v.ManufacturerTokenID.IsZero() {
	//	result.ManufacturerTokenID = v.ManufacturerTokenID.Int
	//}

	if im.AutoPiPowertrainToTemplateID != nil {
		result.AutoPiPowertrainToTemplateID = im.AutoPiPowertrainToTemplateID
	}

	return result, nil

}
