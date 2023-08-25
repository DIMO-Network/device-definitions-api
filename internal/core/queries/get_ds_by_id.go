package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"strings"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type GetDeviceStyleByIDQuery struct {
	DeviceStyleID string `json:"device_style_id"`
}

func (*GetDeviceStyleByIDQuery) Key() string { return "GetDeviceStyleByIDQuery" }

type GetDeviceStyleByIDQueryHandler struct {
	DBS     func() *db.ReaderWriter
	DDCache services.DeviceDefinitionCacheService
}

func NewGetDeviceStyleByIDQueryHandler(dbs func() *db.ReaderWriter, cache services.DeviceDefinitionCacheService) GetDeviceStyleByIDQueryHandler {
	return GetDeviceStyleByIDQueryHandler{DBS: dbs, DDCache: cache}
}

func (ch GetDeviceStyleByIDQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceStyleByIDQuery)

	v, err := models.DeviceStyles(models.DeviceStyleWhere.ID.EQ(qry.DeviceStyleID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device style id: %s", qry.DeviceStyleID),
			}
		}

		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device styles"),
		}
	}

	result := coremodels.GetDeviceStyleQueryResult{
		ID:                 v.ID,
		DeviceDefinitionID: v.DeviceDefinitionID,
		Name:               v.Name,
		ExternalStyleID:    v.ExternalStyleID,
		Source:             v.Source,
		SubModel:           v.SubModel,
	}

	if v.HardwareTemplateID.Valid {
		result.HardwareTemplateID = v.HardwareTemplateID.String
	}

	dd, err := ch.DDCache.GetDeviceDefinitionByID(ctx, result.DeviceDefinitionID)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definition"),
		}
	}

	result.DeviceDefinition = coremodels.GetDeviceDefinitionStyleQueryResult{
		DeviceAttributes: dd.DeviceAttributes,
	}

	// Set default powertrain
	name := strings.ToLower(result.Name)
	powerTrainType := ""
	if strings.Contains(name, "phev") {
		powerTrainType = models.PowertrainPHEV
	} else if strings.Contains(name, "hev") {
		powerTrainType = models.PowertrainHEV
	} else if strings.Contains(name, "plug-in") {
		powerTrainType = models.PowertrainPHEV
	} else if strings.Contains(name, "hybrid") {
		powerTrainType = models.PowertrainHEV
	}

	hasPowertrain := false
	for _, item := range result.DeviceDefinition.DeviceAttributes {
		if item.Name == common.PowerTrainType {
			hasPowertrain = true
			if len(powerTrainType) > 0 {
				item.Value = powerTrainType
			}
			break
		}
	}

	if !hasPowertrain {
		if len(powerTrainType) == 0 {
			powerTrainType = models.PowertrainICE
		}
		result.DeviceDefinition.DeviceAttributes = append(result.DeviceDefinition.DeviceAttributes, coremodels.DeviceTypeAttribute{
			Name:        common.DefaultDeviceType,
			Description: common.DefaultDeviceType,
			Value:       powerTrainType,
		})
	}

	return result, nil
}
