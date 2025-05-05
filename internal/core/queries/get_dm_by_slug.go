package queries

import (
	"context"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/shared/pkg/db"
)

type GetDeviceMakeBySlugQuery struct {
	Slug string `json:"slug"`
}

func (*GetDeviceMakeBySlugQuery) Key() string { return "GetDeviceMakeBySlugQuery" }

type GetDeviceMakeBySlugQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceMakeBySlugQueryHandler(dbs func() *db.ReaderWriter) GetDeviceMakeBySlugQueryHandler {
	return GetDeviceMakeBySlugQueryHandler{DBS: dbs}
}

func (ch GetDeviceMakeBySlugQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetDeviceMakeBySlugQuery)

	dm, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(qry.Slug)).One(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}
	cdm := &coremodels.DeviceMake{
		ID:              dm.ID,
		Name:            dm.Name,
		NameSlug:        dm.NameSlug,
		LogoURL:         dm.LogoURL,
		OemPlatformName: dm.OemPlatformName,
	}

	return cdm, nil
}
