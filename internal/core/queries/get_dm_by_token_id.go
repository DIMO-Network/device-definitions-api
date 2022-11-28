package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"math/big"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type GetDeviceMakeByTokenIDQuery struct {
	TokenID string `json:"tokenId"`
}

func (*GetDeviceMakeByTokenIDQuery) Key() string { return "GetDeviceMakeByTokenIDQuery" }

type GetDeviceMakeByTokenIDQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceMakeByTokenIDQueryHandler(dbs func() *db.ReaderWriter) GetDeviceMakeByTokenIDQueryHandler {
	return GetDeviceMakeByTokenIDQueryHandler{DBS: dbs}
}

func (ch GetDeviceMakeByTokenIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceMakeByTokenIDQuery)

	ti, ok := new(big.Int).SetString(qry.TokenID, 10)
	if !ok {
		return nil, &exceptions.ValidationError{
			Err: fmt.Errorf("Couldn't parse token id."),
		}
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	v, err := models.DeviceMakes(models.DeviceMakeWhere.TokenID.EQ(tid)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make name: %s", qry.TokenID),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}

	result := coremodels.DeviceMake{
		ID:               v.ID,
		Name:             v.Name,
		LogoURL:          v.LogoURL,
		OemPlatformName:  v.OemPlatformName,
		NameSlug:         v.NameSlug,
		ExternalIds:      common.JSONOrDefault(v.ExternalIds),
		ExternalIdsTyped: common.BuildExternalIds(v.ExternalIds),
		Metadata:         common.JSONOrDefault(v.Metadata),
		MetadataTyped:    common.BuildDeviceMakeMetadata(v.Metadata),
	}

	if !v.TokenID.IsZero() {
		result.TokenID = v.TokenID.Big.Int(new(big.Int))
	}

	return result, nil
}
