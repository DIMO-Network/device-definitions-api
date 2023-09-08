package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
)

type GetDeviceDefinitionHardwareTemplateByIDQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	IntegrationID      string `json:"integration_id"`
}

func (*GetDeviceDefinitionHardwareTemplateByIDQuery) Key() string {
	return "GetDeviceDefinitionHardwareTemplateByIdQuery"
}

type GetDeviceDefinitionHardwareTemplateByIDQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceDefinitionHardwareTemplateByIDQueryHandler(cache services.DeviceDefinitionCacheService) GetDeviceDefinitionHardwareTemplateByIDQueryHandler {
	return GetDeviceDefinitionHardwareTemplateByIDQueryHandler{
		DDCache: cache,
	}
}

func (ch GetDeviceDefinitionHardwareTemplateByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionHardwareTemplateByIDQuery)

	dd, err := ch.DDCache.GetDeviceDefinitionByID(ctx, qry.DeviceDefinitionID)

	if err != nil {
		return nil, err
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", qry.DeviceDefinitionID),
		}
	}

	result := models.GetDeviceDefinitionHardwareTemplateQueryResult{}

	if dd.DeviceMake.HardwareTemplateID.Valid && len(dd.DeviceMake.HardwareTemplateID.String) > 0 {
		result.TemplateID = dd.DeviceMake.HardwareTemplateID.String

		return result, nil
	}

	return result, nil
}
