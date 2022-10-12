package queries

import (
	"context"
	"fmt"
	"math/big"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetAllDeviceMakeQuery struct {
	Name string `json:"name"`
}

func (*GetAllDeviceMakeQuery) Key() string { return "GetAllDeviceMakeQuery" }

type GetAllDeviceMakeQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetAllDeviceMakeQueryHandler(dbs func() *db.ReaderWriter) GetAllDeviceMakeQueryHandler {
	return GetAllDeviceMakeQueryHandler{DBS: dbs}
}

func (ch GetAllDeviceMakeQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, err := models.DeviceMakes().All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}

	result := make([]coremodels.DeviceMake, len(all))
	for i, v := range all {
		result[i] = coremodels.DeviceMake{
			ID:              v.ID,
			Name:            v.Name,
			LogoURL:         v.LogoURL,
			OemPlatformName: v.OemPlatformName,
			TokenID:         v.TokenID.Big.Int(new(big.Int)),
			NameSlug:        v.NameSlug,
			ExternalIds:     common.JSONOrDefault(v.ExternalIds),
		}
	}

	return result, nil
}
