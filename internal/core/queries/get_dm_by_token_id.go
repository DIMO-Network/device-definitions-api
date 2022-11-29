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
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
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
			Err: fmt.Errorf("Couldn't parse token id"),
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

	eids := common.BuildExternalIds(v.ExternalIds)
	md := &coremodels.DeviceMakeMetadata{}
	_ = v.Metadata.Unmarshal(md)

	result := &p_grpc.DeviceMake{
		Id:               v.ID,
		Name:             v.Name,
		LogoUrl:          v.LogoURL.String,
		OemPlatformName:  v.OemPlatformName.String,
		NameSlug:         v.NameSlug,
		ExternalIds:      string(v.ExternalIds.JSON),
		ExternalIdsTyped: common.ExternalIdsToGRPC(eids),
		Metadata:         common.DeviceMakeMetadataToGRPC(md),
	}

	if !v.TokenID.IsZero() {
		result.TokenId = v.TokenID.Big.Int(new(big.Int)).Uint64()
	}

	return result, nil
}
