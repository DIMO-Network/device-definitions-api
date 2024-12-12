package queries

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/volatiletech/null/v8"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
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
	ddCache       services.DeviceDefinitionCacheService
}

func NewGetDeviceMakeByTokenIDQueryHandler(dbs func() *db.ReaderWriter, registryInstance *contracts.Registry, ddCache services.DeviceDefinitionCacheService) GetDeviceMakeByTokenIDQueryHandler {
	return GetDeviceMakeByTokenIDQueryHandler{DBS: dbs, queryInstance: registryInstance, ddCache: ddCache}
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
	dm, err := ch.ddCache.GetDeviceMakeByName(ctx, manufName)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device make by name: %s", manufName),
		}
	}

	metadata := common.BuildDeviceMakeMetadata(null.JSONFrom(dm.Metadata))

	result := &p_grpc.DeviceMake{
		Id:              dm.ID,
		Name:            dm.Name,
		LogoUrl:         dm.LogoURL.String,
		OemPlatformName: dm.OemPlatformName.String,
		NameSlug:        dm.NameSlug,
		TokenId:         ti.Uint64(),
		Metadata:        common.DeviceMakeMetadataToGRPC(metadata),
		CreatedAt:       timestamppb.New(dm.CreatedAt),
		UpdatedAt:       timestamppb.New(dm.UpdatedAt),
	}

	return result, nil
}
