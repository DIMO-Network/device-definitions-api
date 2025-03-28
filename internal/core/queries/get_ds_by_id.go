//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
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

	ds, err := models.DeviceStyles(models.DeviceStyleWhere.ID.EQ(qry.DeviceStyleID)).One(ctx, ch.DBS().Reader)
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
	dd, err := ch.DDCache.GetDeviceDefinitionByID(ctx, ds.DefinitionID)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definition"),
		}
	}

	deviceStyleResult := coremodels.GetDeviceStyleQueryResult{
		ID:                 ds.ID,
		DefinitionID:       ds.DefinitionID,
		Name:               ds.Name,
		ExternalStyleID:    ds.ExternalStyleID,
		Source:             ds.Source,
		SubModel:           ds.SubModel,
		HardwareTemplateID: ds.HardwareTemplateID.String,
		DeviceDefinition: coremodels.GetDeviceDefinitionStyleQueryResult{
			DeviceAttributes: dd.DeviceAttributes, // copy any attributes from parent DD
		},
	}
	// first see if style metadata has powertrain, most cases will be blank
	powerTrainType := gjson.GetBytes(ds.Metadata.JSON, common.PowerTrainType).String()
	if len(powerTrainType) == 0 {
		// Set powertrain based on naming logic
		name := strings.ToLower(deviceStyleResult.Name)
		if strings.Contains(name, "phev") {
			powerTrainType = models.PowertrainPHEV
		} else if strings.Contains(name, "hev") {
			powerTrainType = models.PowertrainHEV
		} else if strings.Contains(name, "plug-in") {
			powerTrainType = models.PowertrainPHEV
		} else if strings.Contains(name, "hybrid") {
			powerTrainType = models.PowertrainHEV
		} else if strings.Contains(name, "electric") {
			powerTrainType = models.PowertrainBEV
		} else if strings.Contains(name, "4xe") {
			powerTrainType = models.PowertrainPHEV
		} else if strings.Contains(name, "energi") {
			powerTrainType = models.PowertrainPHEV
		}
	}

	// override any existing powertrain inherited from device definition, only if we came up with something worthy from above logic
	hasPowertrain := false
	for i, item := range deviceStyleResult.DeviceDefinition.DeviceAttributes {
		if item.Name == common.PowerTrainType {
			hasPowertrain = true
			if len(powerTrainType) > 0 {
				deviceStyleResult.DeviceDefinition.DeviceAttributes[i].Value = powerTrainType
			}
			break
		}
	}

	// if no powertrain attribute found, set it, defaulting to parent DD if nothing resulted from above logic
	if !hasPowertrain {
		if len(powerTrainType) == 0 {
			ddPt := gjson.GetBytes(dd.Metadata, common.PowerTrainType).String()
			if len(ddPt) > 0 {
				powerTrainType = ddPt
			} else {
				powerTrainType = models.PowertrainICE // default to ICE if nothing found
			}
		}
		deviceStyleResult.DeviceDefinition.DeviceAttributes = append(deviceStyleResult.DeviceDefinition.DeviceAttributes, coremodels.DeviceTypeAttribute{
			Name:        common.PowerTrainType,
			Description: common.PowerTrainType,
			Type:        common.DefaultDeviceType,
			Value:       powerTrainType,
		})
	}

	return deviceStyleResult, nil
}
