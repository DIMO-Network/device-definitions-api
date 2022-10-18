package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceCompatibilityQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
}

type GetDeviceCompatibilityQueryResult struct {
	DeviceDefinitions   models.DeviceDefinitionSlice
	IntegrationFeatures map[string]map[string]string
}

type GetDeviceCompatibilityQuery struct {
	MakeID        string `json:"makeId" validate:"required"`
	IntegrationID string `json:"integrationId" validate:"required"`
	Region        string `json:"region" validate:"required"`
}

func (*GetDeviceCompatibilityQuery) Key() string { return "GetDeviceCompatibilityQuery" }

func NewGetDeviceCompatibilityQueryHandler(dbs func() *db.ReaderWriter, repository repositories.DeviceDefinitionRepository) GetDeviceCompatibilityQueryHandler {
	return GetDeviceCompatibilityQueryHandler{
		Repository: repository,
		DBS:        dbs,
	}
}

func (dc GetDeviceCompatibilityQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetDeviceCompatibilityQuery)

	inf, err := models.IntegrationFeatures().All(ctx, dc.DBS().Reader)
	if err != nil {
		return GetDeviceCompatibilityQueryResult{}, err
	}

	integFeats := make(map[string]map[string]string, len(inf))
	for _, k := range inf {
		integFeats[k.FeatureKey] = map[string]string{
			"displayName": k.DisplayName,
			"cssIcon":     k.CSSIcon.String,
		}
	}

	res, err := dc.Repository.FetchDeviceCompatibility(ctx, qry.MakeID, qry.IntegrationID, qry.Region)
	if err != nil {
		return GetDeviceCompatibilityQueryResult{}, err
	}

	return GetDeviceCompatibilityQueryResult{
		DeviceDefinitions:   res,
		IntegrationFeatures: integFeats,
	}, nil
}
