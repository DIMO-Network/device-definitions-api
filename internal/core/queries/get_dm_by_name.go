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

type GetDeviceMakeByNameQuery struct {
	Name string `json:"name"`
}

func (*GetDeviceMakeByNameQuery) Key() string { return "GetDeviceMakeByNameQuery" }

type GetDeviceMakeByNameQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceMakeByNameQueryHandler(dbs func() *db.ReaderWriter) GetDeviceMakeByNameQueryHandler {
	return GetDeviceMakeByNameQueryHandler{DBS: dbs}
}

func (ch GetDeviceMakeByNameQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceMakeByNameQuery)

	v, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(qry.Name)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make name: %s", qry.Name),
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
		TokenID:         v.TokenID.Big.Int(new(big.Int)),
		NameSlug:        v.NameSlug,
		ExternalIds:     common.JSONOrDefault(v.ExternalIds),
	}

	return result, nil
}
