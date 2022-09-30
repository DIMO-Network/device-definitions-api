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
	Model    string
	Year     int32
	Features map[string]interface{}
}

type GetDeviceCompatibilityQuery struct {
	MakeID string `json:"makeId" validate:"required"`
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
		return GetDeviceCompatibilityQueryResult{}, nil
	}
	integFeats := make(map[string]string, len(inf))
	for _, k := range inf {
		integFeats[k.FeatureKey] = k.DisplayName
	}

	res, err := dc.Repository.FetchCompatibilityByMakeID(ctx, qry.MakeID)
	if err != nil {
		return GetDeviceCompatibilityQueryResult{}, nil
	}

	var resp []GetDeviceCompatibilityQueryResult
	for _, v := range res {
		if v.DeviceIntegration.Features.IsZero() {
			continue
		}
		cr := GetDeviceCompatibilityQueryResult{
			Model: v.DeviceDefinition.Model,
			Year:  int32(v.DeviceDefinition.Year),
		}
		var dd []interface{}
		feats := make(map[string]interface{})
		err = v.DeviceIntegration.Features.Unmarshal(&dd)
		if err != nil {
			return GetDeviceCompatibilityQueryResult{}, nil
		}
		for _, i := range dd {
			f := i.(map[string]interface{})
			fk := f["feature_key"]
			sl := f["support_level"]
			feats[integFeats[fk.(string)]] = sl.(float64) > 0
		}

		cr.Features = feats

		resp = append(resp, cr)
	}

	return resp, nil
}
