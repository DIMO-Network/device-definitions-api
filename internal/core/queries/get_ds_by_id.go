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

type GetDeviceStyleByIDQuery struct {
	DeviceStyleID string `json:"device_style_id"`
}

func (*GetDeviceStyleByIDQuery) Key() string { return "GetDeviceStyleByIDQuery" }

type GetDeviceStyleByIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceStyleByIDQueryHandler(dbs func() *db.ReaderWriter) GetDeviceStyleByIDQueryHandler {
	return GetDeviceStyleByIDQueryHandler{DBS: dbs}
}

func (ch GetDeviceStyleByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceStyleByIDQuery)

	v, err := models.DeviceStyles(models.DeviceStyleWhere.ID.EQ(qry.DeviceStyleID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device style id: %s", qry.DeviceStyleID),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device styles"),
		}
	}

	result := coremodels.GetDeviceStyleQueryResult{
		ID:                 v.ID,
		DeviceDefinitionID: v.DeviceDefinitionID,
		Name:               v.Name,
		ExternalStyleID:    v.ExternalStyleID,
		Source:             v.Source,
		SubModel:           v.SubModel,
	}

	return result, nil
}
