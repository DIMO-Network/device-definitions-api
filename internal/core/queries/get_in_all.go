package queries

import (
	"context"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetAllIntegrationQuery struct {
}

type GetAllIntegrationQueryResult struct {
	ID                           string                            `json:"id"`
	Type                         string                            `json:"type"`
	Style                        string                            `json:"style"`
	Vendor                       string                            `json:"vendor"`
	AutoPiDefaultTemplateID      int                               `json:"autoPiDefaultTemplateId"`
	AutoPiPowertrainToTemplateID map[coremodels.PowertrainType]int `json:"autoPiPowertrainToTemplateId,omitempty"`
}

func (*GetAllIntegrationQuery) Key() string { return "GetAllIntegrationQuery" }

type GetAllIntegrationQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetAllIntegrationQueryHandler(dbs func() *db.ReaderWriter) GetAllIntegrationQueryHandler {
	return GetAllIntegrationQueryHandler{DBS: dbs}
}

func (ch GetAllIntegrationQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, err := models.Integrations(qm.Limit(100)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integrations"),
		}
	}

	result := make([]GetAllIntegrationQueryResult, len(all))
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
		result[i] = GetAllIntegrationQueryResult{
			ID:                      v.ID,
			Type:                    v.Type,
			Style:                   v.Style,
			Vendor:                  v.Vendor,
			AutoPiDefaultTemplateID: im.AutoPiDefaultTemplateID,
		}
		if im.AutoPiPowertrainToTemplateID != nil {
			result[i].AutoPiPowertrainToTemplateID = im.AutoPiPowertrainToTemplateID
		}
	}

	return result, nil

}
