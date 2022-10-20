package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

const (
	GOLD_LEVEL   = "Gold"
	SILVER_LEVEL = "Silver"
	BRONZE_LEVEL = "Bronze"
)

type GetDeviceCompatibilityQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
}

type FeatureDetails struct {
	DisplayName   string
	CSSIcon       string
	FeatureWeight float64
	SupportLevel  int32
}

type GetDeviceCompatibilityQueryResult struct {
	DeviceDefinitions   models.DeviceDefinitionSlice
	IntegrationFeatures map[string]FeatureDetails
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

func GetDeviceCompatibilityLevel(fd map[string]FeatureDetails, totalWeights float64) string {
	level := ""
	total := 0.0

	for _, v := range fd {
		if v.SupportLevel > 0 {
			total += v.FeatureWeight
		} else {
			total += 0
		}
	}

	if total != 0 && total < totalWeights {
		p := (total / totalWeights) * 100

		if p >= 75 {
			level = GOLD_LEVEL
		} else if p > 50 && p < 75 {
			level = SILVER_LEVEL
		} else {
			level = BRONZE_LEVEL
		}
	}

	return level
}

func (dc GetDeviceCompatibilityQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetDeviceCompatibilityQuery)

	inf, err := models.IntegrationFeatures().All(ctx, dc.DBS().Reader)
	if err != nil {
		return GetDeviceCompatibilityQueryResult{}, err
	}

	integFeats := make(map[string]FeatureDetails, len(inf))
	for _, k := range inf {
		integFeats[k.FeatureKey] = FeatureDetails{
			DisplayName:   k.DisplayName,
			CSSIcon:       k.CSSIcon.String,
			FeatureWeight: k.FeatureWeight.Float64,
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
