package queries

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
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

	v, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(qry.Slug)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make slug: %s", qry.Slug),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}

	result := coremodels.DeviceMake{
		ID:              v.ID,
		Name:            v.Name,
		LogoURL:         v.LogoURL,
		OemPlatformName: v.OemPlatformName,
		NameSlug:        v.NameSlug,
		ExternalIds:     common.JSONOrDefault(v.ExternalIds),
	}

	if !v.TokenID.IsZero() {
		result.TokenID = v.TokenID.Big.Int(new(big.Int))
	}

	return result, nil
}
