package queries

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
)

type GetDeviceMakeByTokenIDQuery struct {
	TokenID string `json:"tokenId"`
}

func (*GetDeviceMakeByTokenIDQuery) Key() string { return "GetDeviceMakeByTokenIDQuery" }

type GetDeviceMakeByTokenIDQueryHandler struct {
	DBS           func() *db.ReaderWriter
	queryInstance *contracts.Registry
}

func NewGetDeviceMakeByTokenIDQueryHandler(dbs func() *db.ReaderWriter, registryInstance *contracts.Registry) GetDeviceMakeByTokenIDQueryHandler {
	return GetDeviceMakeByTokenIDQueryHandler{DBS: dbs, queryInstance: registryInstance}
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

	manufName, err := ch.queryInstance.GetManufacturerNameById(&bind.CallOpts{Context: ctx, Pending: true}, ti)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get manufacturer name by token id: %s", qry.TokenID),
		}
	}

	dm, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(manufName)).One(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}

	result := &p_grpc.DeviceMake{
		Id:              dm.ID,
		Name:            dm.Name,
		LogoUrl:         dm.LogoURL.String,
		OemPlatformName: dm.OemPlatformName.String,
		NameSlug:        dm.NameSlug,
		TokenId:         ti.Uint64(),
		CreatedAt:       timestamppb.New(dm.CreatedAt),
		UpdatedAt:       timestamppb.New(dm.UpdatedAt),
	}

	return result, nil
}
