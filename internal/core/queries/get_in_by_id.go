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

type GetIntegrationByIDQuery struct {
	IntegrationID string `json:"integration_id"`
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

	v, err := models.Integrations(models.IntegrationWhere.ID.EQ(qry.IntegrationID)).One(ctx, ch.DBS().Reader)
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
	}

	if im.AutoPiPowertrainToTemplateID != nil {
		result.AutoPiPowertrainToTemplateID = im.AutoPiPowertrainToTemplateID
	}

	return result, nil

}
