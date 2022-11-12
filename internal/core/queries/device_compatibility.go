package queries

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

// CompatibilityLevel enum for overall device compatibility
type CompatibilityLevel string

const (
	GoldLevel   CompatibilityLevel = "Gold"
	SilverLevel CompatibilityLevel = "Silver"
	BronzeLevel CompatibilityLevel = "Bronze"
	NoDataLevel CompatibilityLevel = "No Data"
)

func (r CompatibilityLevel) String() string {
	return string(r)
}

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
	Cursor        string `json:"cursor"`
	Size          int64  `json:"size"`
}

func (*GetDeviceCompatibilityQuery) Key() string { return "GetDeviceCompatibilityQuery" }

func NewGetDeviceCompatibilityQueryHandler(dbs func() *db.ReaderWriter, repository repositories.DeviceDefinitionRepository) GetDeviceCompatibilityQueryHandler {
	return GetDeviceCompatibilityQueryHandler{
		Repository: repository,
		DBS:        dbs,
	}
}

func GetDeviceCompatibilityLevel(fd map[string]FeatureDetails, totalWeights float64) CompatibilityLevel {
	featureWeight := 0.0

	for _, v := range fd {
		if v.SupportLevel > 0 {
			featureWeight += v.FeatureWeight
		}
	}
	return calculateMathForLevel(featureWeight, totalWeights)
}

// calculateMathForLevel does the math to figure out compatibility level based on sum of all weights and total weights of all available features
func calculateMathForLevel(featuresWeight, totalWeights float64) CompatibilityLevel {
	if featuresWeight != 0 && featuresWeight <= totalWeights {
		p := (featuresWeight / totalWeights) * 100
		if p >= 75 {
			return GoldLevel
		} else if p >= 50 {
			return SilverLevel
		} else {
			return BronzeLevel
		}
	}
	return NoDataLevel
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

	res, err := dc.Repository.FetchDeviceCompatibility(ctx, qry.MakeID, qry.IntegrationID, qry.Region, qry.Cursor, qry.Size)
	if err != nil {
		return GetDeviceCompatibilityQueryResult{}, err
	}

	return GetDeviceCompatibilityQueryResult{
		DeviceDefinitions:   res,
		IntegrationFeatures: integFeats,
	}, nil
}
