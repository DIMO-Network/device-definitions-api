package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetCompatibilityByDeviceDefinitionArrayQuery struct {
	DeviceDefinitionID []string
}

type GetCompatibilityByDeviceDefinitionArrayItem struct {
	DeviceDefinitionID string
}

func (*GetCompatibilityByDeviceDefinitionArrayQuery) Key() string {
	return "GetCompatibilityByDeviceDefinitionArrayQuery"
}

type GetCompatibilityByDeviceDefinitionArrayQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetCompatibilityByDeviceDefinitionArrayQueryHandler(dbs func() *db.ReaderWriter) GetCompatibilityByDeviceDefinitionArrayQueryHandler {
	return GetCompatibilityByDeviceDefinitionArrayQueryHandler{DBS: dbs}
}

func (ch GetCompatibilityByDeviceDefinitionArrayQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetCompatibilityByDeviceDefinitionArrayQuery)
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.IN(qry.DeviceDefinitionID)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}
	result := &grpc.GetCompatibilityByDeviceArrayResponse{}
	for _, device := range dd {
		dd, err := models.FindDeviceDefinition(ctx, ch.DBS().Reader, device.ID)
		if err != nil {
			return nil, &exceptions.InternalError{Err: err}
		}
		// todo will need the powertrain from the metadata for non Tesla, Rivian and Lucid makes. use the dd cache service?
		integFeats, totalWeights, err := getIntegrationFeatures(ctx, dd.DeviceMakeID, ch.DBS().Reader)
		if err != nil {
			return nil, &exceptions.InternalError{Err: err}
		}

		response := &p_grpc.GetCompatibilityByDeviceArrayResponseItem{}
		dis, err := models.DeviceIntegrations(
			qm.Load(models.DeviceIntegrationRels.Integration),
			models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(device.ID)).All(ctx, ch.DBS().Reader)
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: errors.Wrap(err, "failed to get device_integrations"),
			}
		}
		response.DeviceDefinitionId = dd.ID
		response.Compatibilities = make([]*p_grpc.DeviceCompatibilities, len(dis))
		for i, di := range dis {
			if di.R.Integration == nil {
				return nil, &exceptions.ConflictError{Err: fmt.Errorf("integration not set or found")}
			}
			gfs := buildFeatures(di.Features, integFeats)
			level, score := calculateCompatibilityLevel(gfs, integFeats, totalWeights)
			response.Compatibilities[i] = &p_grpc.DeviceCompatibilities{
				IntegrationId:     di.IntegrationID,
				IntegrationVendor: di.R.Integration.Vendor,
				Region:            di.Region,
				Features:          gfs,
				Level:             level.String(),
				Score:             float32(score),
			}
		}

		result.Response = append(result.Response, response)
	}

	return result, nil
}
