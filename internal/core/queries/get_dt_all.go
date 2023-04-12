package queries

import (
	"context"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetAllDeviceTypeQuery struct {
}

func (*GetAllDeviceTypeQuery) Key() string { return "GetAllDeviceTypeQuery" }

type GetAllDeviceTypeQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetAllDeviceTypeQueryHandler(dbs func() *db.ReaderWriter) GetAllDeviceTypeQueryHandler {
	return GetAllDeviceTypeQueryHandler{DBS: dbs}
}

func (ch GetAllDeviceTypeQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {

	all, err := models.DeviceTypes().All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device types"),
		}
	}

	result := make([]coremodels.GetDeviceTypeQueryResult, len(all))
	for i, v := range all {
		result[i] = coremodels.GetDeviceTypeQueryResult{
			ID:          v.ID,
			Name:        v.Name,
			Metadatakey: v.Metadatakey,
		}

		// attribute info
		var ai map[string][]coremodels.GetDeviceTypeAttributeQueryResult
		if err := v.Properties.Unmarshal(&ai); err == nil {
			result[i].Attributes = append(result[i].Attributes, ai["properties"]...)
		}
	}

	return result, nil
}
