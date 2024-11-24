package queries

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
)

type GetAllDeviceMakeQuery struct {
}

func (*GetAllDeviceMakeQuery) Key() string { return "GetAllDeviceMakeQuery" }

type GetAllDeviceMakeQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetAllDeviceMakeQueryHandler(dbs func() *db.ReaderWriter) GetAllDeviceMakeQueryHandler {
	return GetAllDeviceMakeQueryHandler{DBS: dbs}
}

func (ch GetAllDeviceMakeQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {

	all, err := models.DeviceMakes().All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}
	result := &p_grpc.GetDeviceMakeResponse{
		DeviceMakes: make([]*p_grpc.DeviceMake, len(all)),
	}

	for i, v := range all {
		md := &coremodels.DeviceMakeMetadata{}
		_ = v.Metadata.Unmarshal(md)

		result.DeviceMakes[i] = &p_grpc.DeviceMake{
			Id:              v.ID,
			Name:            v.Name,
			LogoUrl:         v.LogoURL.String,
			OemPlatformName: v.OemPlatformName.String,
			NameSlug:        v.NameSlug,
			Metadata:        common.DeviceMakeMetadataToGRPC(md),
			CreatedAt:       timestamppb.New(v.CreatedAt),
			UpdatedAt:       timestamppb.New(v.UpdatedAt),
		}

		if !v.TokenID.IsZero() {
			result.DeviceMakes[i].TokenId = v.TokenID.Big.Int(new(big.Int)).Uint64()
		}

	}

	return result, nil
}
