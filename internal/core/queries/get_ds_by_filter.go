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

type GetDeviceStyleByFilterQuery struct {
	DeviceDefinitionID string `json:"device_definition_id"`
	Name               string `json:"name"`
	SubModel           string `json:"sub_model"`
}

func (*GetDeviceStyleByFilterQuery) Key() string {
	return "GetDeviceStyleByFilterQuery"
}

type GetDeviceStyleByFilterQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceStyleByFilterQueryHandler(dbs func() *db.ReaderWriter) GetDeviceStyleByFilterQueryHandler {
	return GetDeviceStyleByFilterQueryHandler{DBS: dbs}
}

func (ch GetDeviceStyleByFilterQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceStyleByFilterQuery)

	style, err := models.DeviceStyles(
		models.DeviceStyleWhere.DeviceDefinitionID.EQ(qry.DeviceDefinitionID),
		models.DeviceStyleWhere.Name.EQ(qry.Name),
		models.DeviceStyleWhere.SubModel.EQ(qry.SubModel),
	).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to get device styles"),
			}
		}
	}

	response := []coremodels.GetDeviceStyleQueryResult{}

	if style == nil {
		return response, nil
	}

	deviceStyle := coremodels.GetDeviceStyleQueryResult{
		ID:                 style.ID,
		DeviceDefinitionID: style.DeviceDefinitionID,
		Name:               style.Name,
		ExternalStyleID:    style.ExternalStyleID,
		Source:             style.Source,
		SubModel:           style.SubModel,
		HardwareTemplateID: style.HardwareTemplateID.String,
	}

	if style.HardwareTemplateID.Valid {
		deviceStyle.HardwareTemplateID = style.HardwareTemplateID.String
	}

	response = append(response, deviceStyle)

	return response, nil
}
