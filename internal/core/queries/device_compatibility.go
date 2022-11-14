package queries

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetDeviceCompatibilityQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
}

type GetCompatibilitiesByMakeQueryResult struct {
	DeviceDefinitions   models.DeviceDefinitionSlice
	IntegrationFeatures map[string]FeatureDetails
}

type GetCompatibilitiesByMakeQuery struct {
	MakeID        string `json:"makeId" validate:"required"`
	IntegrationID string `json:"integrationId" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Cursor        string `json:"cursor"`
	Size          int64  `json:"size"`
}

func (*GetCompatibilitiesByMakeQuery) Key() string { return "GetCompatibilitiesByMakeQuery" }

func NewGetDeviceCompatibilityQueryHandler(dbs func() *db.ReaderWriter, repository repositories.DeviceDefinitionRepository) GetDeviceCompatibilityQueryHandler {
	return GetDeviceCompatibilityQueryHandler{
		Repository: repository,
		DBS:        dbs,
	}
}

func (dc GetDeviceCompatibilityQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetCompatibilitiesByMakeQuery)
	integFeats, err := models.IntegrationFeatures(qm.Limit(100)).All(ctx, dc.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration_features"),
		}
	}
	// todo review what this is doing.
	// I think I should just do what i do in GetCompatibilityByDeviceDefinitionQueryHandler but by make
	// order by desc year, model slug
	res, err := dc.Repository.FetchDeviceCompatibility(ctx, qry.MakeID, qry.IntegrationID, qry.Region, qry.Cursor, qry.Size)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device_integrations"),
		}
	}
	for i, re := range res {

	}

	return &p_grpc.GetCompatibilitiesByMakeResponse{
		Models: nil,
		Cursor: "",
	}, nil
}

// todo delete this when ready
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

type FeatureDetails struct {
	DisplayName   string
	CSSIcon       string
	FeatureWeight float64
	SupportLevel  int32
}

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
