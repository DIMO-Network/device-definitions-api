package queries

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/ericlagergren/decimal"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
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
	qry.TokenID = strings.TrimSpace(qry.TokenID)

	ti, ok := new(big.Int).SetString(qry.TokenID, 10)
	if !ok {
		return nil, &exceptions.ValidationError{
			Err: fmt.Errorf("couldn't parse token id"),
		}
	}

	tid := types.NewNullDecimal(new(decimal.Big).SetBigMantScale(ti, 0))

	v, err := models.DeviceMakes(qm.Where("token_id = ?", qry.TokenID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device make with tokenId param: %s, and bigint tid %v", qry.TokenID, tid),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device make by token id"),
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

	if v.HardwareTemplateID.Valid {
		result.HardwareTemplateId = v.HardwareTemplateID.String
	}

	return result, nil
}
