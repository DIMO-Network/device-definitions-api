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

type GetDeviceStyleByExternalIDQuery struct {
	ExternalDeviceID string `json:"external_device_id"`
}

func (*GetDeviceStyleByExternalIDQuery) Key() string { return "GetDeviceStyleByExternalIDQuery" }

type GetDeviceStyleByExternalIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceStyleByExternalIDQueryHandler(dbs func() *db.ReaderWriter) GetDeviceStyleByExternalIDQueryHandler {
	return GetDeviceStyleByExternalIDQueryHandler{DBS: dbs}
}

func (ch GetDeviceStyleByExternalIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceStyleByExternalIDQuery)

	v, err := models.DeviceStyles(models.DeviceStyleWhere.ExternalStyleID.EQ(qry.ExternalDeviceID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device style external id: %s", qry.ExternalDeviceID),
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
