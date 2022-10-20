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

type GetDeviceTypeByIDQuery struct {
	DeviceTypeID string `json:"device_type_id"`
}

func (*GetDeviceTypeByIDQuery) Key() string { return "GetDeviceTypeByIDQuery" }

type GetDeviceTypeByIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceTypeByIDQueryHandler(dbs func() *db.ReaderWriter) GetDeviceTypeByIDQueryHandler {
	return GetDeviceTypeByIDQueryHandler{DBS: dbs}
}

func (ch GetDeviceTypeByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceTypeByIDQuery)

	v, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(qry.DeviceTypeID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device type id: %s", qry.DeviceTypeID),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device types"),
		}
	}

	result := coremodels.GetDeviceTypeQueryResult{
		ID:   v.ID,
		Name: v.Name,
	}

	// attribute info
	var ai map[string][]coremodels.GetDeviceTypeAttributeQueryResult
	if err := v.Properties.Unmarshal(&ai); err == nil {
		result.Attributes = append(result.Attributes, ai["properties"]...)
	}

	return result, nil
}
