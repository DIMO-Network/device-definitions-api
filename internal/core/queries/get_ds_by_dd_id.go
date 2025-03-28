//nolint:tagliatelle
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

type GetDeviceStyleByDeviceDefinitionIDQuery struct {
	DefinitionID string `json:"definition_id"`
}

func (*GetDeviceStyleByDeviceDefinitionIDQuery) Key() string {
	return "GetDeviceStyleByDeviceDefinitionIDQuery"
}

type GetDeviceStyleByDeviceDefinitionIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceStyleByDeviceDefinitionIDQueryHandler(dbs func() *db.ReaderWriter) GetDeviceStyleByDeviceDefinitionIDQueryHandler {
	return GetDeviceStyleByDeviceDefinitionIDQueryHandler{DBS: dbs}
}

func (ch GetDeviceStyleByDeviceDefinitionIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceStyleByDeviceDefinitionIDQuery)

	styles, err := models.DeviceStyles(models.DeviceStyleWhere.DefinitionID.EQ(qry.DefinitionID)).All(ctx, ch.DBS().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to get device styles"),
			}
		}
	}

	response := []coremodels.GetDeviceStyleQueryResult{}

	for _, v := range styles {
		deviceStyle := coremodels.GetDeviceStyleQueryResult{
			ID:              v.ID,
			DefinitionID:    v.DefinitionID,
			Name:            v.Name,
			ExternalStyleID: v.ExternalStyleID,
			Source:          v.Source,
			SubModel:        v.SubModel,
		}

		if v.HardwareTemplateID.Valid {
			deviceStyle.HardwareTemplateID = v.HardwareTemplateID.String
		}

		response = append(response, deviceStyle)
	}

	return response, nil
}
