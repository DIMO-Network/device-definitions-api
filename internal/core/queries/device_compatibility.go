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

type Feature struct {
	Key          string // eg. odometer
	SupportLevel int    // eg. 0,1,2
}

type GetDeviceCompatibilityQueryResult struct {
	Model    string
	Year     int32
	Features []Feature
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
		return GetDeviceCompatibilityQueryResult{}, err
	}

	integFeats := make(map[string]string, len(inf))
	for _, k := range inf {
		integFeats[k.FeatureKey] = k.DisplayName
	}

	res, err := dc.Repository.FetchCompatibilityByMakeID(ctx, qry.MakeID)
	if err != nil {
		return GetDeviceCompatibilityQueryResult{}, err
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
		feats := []Feature{}
		if v.DeviceIntegration.Features.IsZero() {
			continue
		}
		err = v.DeviceIntegration.Features.Unmarshal(&dd)
		if err != nil {
			return GetDeviceCompatibilityQueryResult{}, nil
		}
		for _, i := range dd {
			f := i.(map[string]interface{})
			ft := &Feature{}
			ft.Key = f["feature_key"].(string)
			ft.SupportLevel = int(f["support_level"].(float64))
		}

		cr.Features = feats

		resp = append(resp, cr)
	}

	return resp, nil
}
