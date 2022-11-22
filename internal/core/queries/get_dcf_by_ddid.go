package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetCompatibilityByDeviceDefinitionQuery struct {
	DeviceDefinitionID string
}

func (*GetCompatibilityByDeviceDefinitionQuery) Key() string {
	return "GetCompatibilityByDeviceDefinitionQuery"
}

type GetCompatibilityByDeviceDefinitionQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetCompatibilityByDeviceDefinitionQueryHandler(dbs func() *db.ReaderWriter) GetCompatibilityByDeviceDefinitionQueryHandler {
	return GetCompatibilityByDeviceDefinitionQueryHandler{DBS: dbs}
}

func (ch GetCompatibilityByDeviceDefinitionQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetCompatibilityByDeviceDefinitionQuery)
	dd, err := models.FindDeviceDefinition(ctx, ch.DBS().Reader, qry.DeviceDefinitionID)
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}
	// todo will need the powertrain from the metadata for non Tesla, Rivian and Lucid makes. use the dd cache service?
	integFeats, totalWeights, err := getIntegrationFeatures(ctx, dd.DeviceMakeID, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}

	response := &p_grpc.GetDeviceCompatibilitiesResponse{}
	dis, err := models.DeviceIntegrations(
		qm.Load(models.DeviceIntegrationRels.Integration),
		models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(qry.DeviceDefinitionID)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: errors.Wrap(err, "failed to get device_integrations"),
		}
	}
	response.Compatibilities = make([]*p_grpc.DeviceCompatibilities, len(dis))
	for i, di := range dis {
		if di.R.Integration == nil {
			return nil, &exceptions.ConflictError{Err: fmt.Errorf("integration not set or found")}
		}
		gfs := buildFeatures(di.Features, integFeats)
		response.Compatibilities[i] = &p_grpc.DeviceCompatibilities{
			IntegrationId:     di.IntegrationID,
			IntegrationVendor: di.R.Integration.Vendor,
			Region:            di.Region,
			Features:          gfs,
			Level:             calculateCompatibilityLevel(gfs, integFeats, totalWeights).String(),
		}
	}
	// build up grpc object
	return response, nil
}
