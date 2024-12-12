package queries

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/shared/db"
)

type GetDeviceMakeByNameQuery struct {
	Name string `json:"name"`
}

func (*GetDeviceMakeByNameQuery) Key() string { return "GetDeviceMakeByNameQuery" }

type GetDeviceMakeByNameQueryHandler struct {
	DBS     func() *db.ReaderWriter
	ddCache services.DeviceDefinitionCacheService
}

func NewGetDeviceMakeByNameQueryHandler(dbs func() *db.ReaderWriter, ddCache services.DeviceDefinitionCacheService) GetDeviceMakeByNameQueryHandler {
	return GetDeviceMakeByNameQueryHandler{DBS: dbs, ddCache: ddCache}
}

func (ch GetDeviceMakeByNameQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetDeviceMakeByNameQuery)
	return ch.ddCache.GetDeviceMakeByName(ctx, qry.Name)
}
