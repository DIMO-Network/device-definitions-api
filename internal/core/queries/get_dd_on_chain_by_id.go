package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionOnChainByIDQuery struct {
	MakeSlug           string `json:"makeSlug"`
	DeviceDefinitionID string `json:"deviceDefinitionId"`
}

func (*GetDeviceDefinitionOnChainByIDQuery) Key() string {
	return "GetDeviceDefinitionOnChainByIDQuery"
}

type GetDeviceDefinitionOnChainByIDQueryHandler struct {
	DBS     func() *db.ReaderWriter
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceDefinitionOnChainByIDQueryHandler(cache services.DeviceDefinitionCacheService, dbs func() *db.ReaderWriter) GetDeviceDefinitionOnChainByIDQueryHandler {
	return GetDeviceDefinitionOnChainByIDQueryHandler{
		DDCache: cache,
		DBS:     dbs,
	}
}

func (ch GetDeviceDefinitionOnChainByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionOnChainByIDQuery)

	make, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(qry.MakeSlug)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make slug: %s", qry.MakeSlug),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}

	dd, err := ch.DDCache.GetDeviceDefinitionByID(ctx, qry.DeviceDefinitionID, services.UseOnChain(make.TokenID))

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	return dd, nil
}
