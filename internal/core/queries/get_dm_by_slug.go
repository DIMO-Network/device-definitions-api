package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/shared/db"
)

type GetDeviceMakeBySlugQuery struct {
	Slug string `json:"slug"`
}

func (*GetDeviceMakeBySlugQuery) Key() string { return "GetDeviceMakeBySlugQuery" }

type GetDeviceMakeBySlugQueryHandler struct {
	DBS     func() *db.ReaderWriter
	ddCache services.DeviceDefinitionCacheService
}

func NewGetDeviceMakeBySlugQueryHandler(dbs func() *db.ReaderWriter, ddCache services.DeviceDefinitionCacheService) GetDeviceMakeBySlugQueryHandler {
	return GetDeviceMakeBySlugQueryHandler{DBS: dbs, ddCache: ddCache}
}

func (ch GetDeviceMakeBySlugQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceMakeBySlugQuery)

	return ch.ddCache.GetDeviceMakeByName(ctx, qry.Slug) // this method does a to-slug for lookup anyways
}
