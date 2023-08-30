package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

type GetIntegrationByTokenIDQuery struct {
	TokenID int `json:"tokenId" validate:"required"`
}

func (*GetIntegrationByTokenIDQuery) Key() string {
	return "GetIntegrationByTokenIDQuery"
}

type GetIntegrationByTokenIDQueryHandler struct {
	DBS func() *db.ReaderWriter
	log *zerolog.Logger
}

func NewGetIntegrationByTokenIDQueryHandler(dbs func() *db.ReaderWriter, log *zerolog.Logger) GetIntegrationByTokenIDQueryHandler {
	return GetIntegrationByTokenIDQueryHandler{
		DBS: dbs,
		log: log,
	}
}

func (ch GetIntegrationByTokenIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetIntegrationByTokenIDQuery)

	integration, err := models.Integrations(
		models.IntegrationWhere.TokenID.EQ(null.IntFrom(qry.TokenID)),
	).One(ctx, ch.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find integration with tokenID: %d ", qry.TokenID),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration using tokenID: %d", qry.TokenID),
		}
	}

	response := coremodels.GetIntegrationQueryResult{}
	im := new(coremodels.IntegrationsMetadata)
	if integration.Metadata.Valid {
		err = integration.Metadata.Unmarshal(&im)

		if err != nil {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to unmarshall integration metadata id %s", integration.ID),
			}
		}
	}

	response = coremodels.GetIntegrationQueryResult{
		ID:                      integration.ID,
		Type:                    integration.Type,
		Style:                   integration.Style,
		Vendor:                  integration.Vendor,
		AutoPiDefaultTemplateID: im.AutoPiDefaultTemplateID,
		RefreshLimitSecs:        integration.RefreshLimitSecs,
		TokenID:                 integration.TokenID.Int,
	}

	if im.AutoPiPowertrainToTemplateID != nil {
		response.AutoPiPowertrainToTemplateID = im.AutoPiPowertrainToTemplateID
	}
	return response, nil
}
