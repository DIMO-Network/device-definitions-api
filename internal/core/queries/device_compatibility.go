package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceCompatibilityQueryHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
}

type GetCompatibilitiesByMakeQuery struct {
	MakeID        string `json:"makeId" validate:"required"`
	IntegrationID string `json:"integrationId" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Skip          int64  `json:"skip"`
	Take          int64  `json:"take"`
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
	if qry.Take == 0 {
		qry.Take = 50
	}
	const columns = 6 // number of columns to get, highest weighted first
	const cutoffYear = 2008

	integFeats, totalWeights, err := getIntegrationFeatures(ctx, dc.DBS().Reader)
	if err != nil {
		return nil, err
	}
	dis, err := models.DeviceIntegrations(
		qm.InnerJoin("integrations i on i.id = device_integrations.integration_id"),
		qm.InnerJoin("device_definitions dd on dd.id = device_integrations.device_definition_id"),
		qm.Where("dd.device_make_id = ?", qry.MakeID),
		models.DeviceIntegrationWhere.IntegrationID.EQ(qry.IntegrationID),
		models.DeviceIntegrationWhere.Region.EQ(qry.Region),
		qm.And("dd.year > ?", cutoffYear),
		qm.Load(models.DeviceIntegrationRels.DeviceDefinition),
		qm.Load(models.DeviceIntegrationRels.Integration),
		qm.OrderBy("(features IS NOT NULL) desc, dd.year DESC, dd.model_slug ASC"), // optimal & fast sorting, but breaks ability to use dd.id as cursor
		qm.Offset(int(qry.Skip)),                                                   // doing regular paging since cursor breaks with current sorting setup
		qm.Limit(int(qry.Take))).
		All(ctx, dc.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: errors.Wrapf(err, "failed to get device_integrations by makeId: %s", qry.MakeID),
		}
	}
	if len(dis) == 0 {
		return &p_grpc.GetCompatibilitiesByMakeResponse{}, nil
	}
	// get the total count for pagination. future: cache this count
	count, err := models.DeviceIntegrations(
		qm.InnerJoin("device_definitions dd on dd.id = device_integrations.device_definition_id"),
		qm.Where("dd.device_make_id = ?", qry.MakeID),
		models.DeviceIntegrationWhere.IntegrationID.EQ(qry.IntegrationID),
		models.DeviceIntegrationWhere.Region.EQ(qry.Region),
		qm.And("dd.year > ?", cutoffYear),
	).Count(ctx, dc.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: errors.Wrapf(err, "failed to get count for device_integrations by makeId: %s", qry.MakeID),
		}
	}

	var modelCompats = make([]*p_grpc.DeviceCompatibilities, len(dis))
	for i, di := range dis {
		gfs := buildFeatures(di.Features, integFeats)
		var reduced []*p_grpc.Feature
		if len(gfs) > 6 {
			reduced = gfs[:columns]
		}
		modelCompats[i] = &p_grpc.DeviceCompatibilities{
			Year:              int32(di.R.DeviceDefinition.Year),
			Features:          reduced,
			Level:             calculateCompatibilityLevel(gfs, integFeats, totalWeights).String(),
			IntegrationId:     di.IntegrationID,
			IntegrationVendor: di.R.Integration.Vendor,
			Region:            di.Region,
			Model:             di.R.DeviceDefinition.Model,
			ModelSlug:         di.R.DeviceDefinition.ModelSlug,
		}
	}

	return &p_grpc.GetCompatibilitiesByMakeResponse{
		Models:     modelCompats,
		TotalCount: count,
	}, nil
}

// getIntegrationFeatures refactos out calling db and getting total weights for all integration features
func getIntegrationFeatures(ctx context.Context, dc *db.DB) (models.IntegrationFeatureSlice, float64, error) {
	// todo cache this
	integFeats, err := models.IntegrationFeatures(qm.OrderBy("feature_weight DESC, feature_key"), qm.Limit(50)).All(ctx, dc)
	if err != nil {
		return nil, 0, &exceptions.InternalError{
			Err: errors.Wrap(err, "failed to get integration_features"),
		}
	}
	totalWeights := 0.0
	for _, v := range integFeats {
		if !v.FeatureWeight.IsZero() {
			totalWeights += v.FeatureWeight.Float64
		}
	}
	return integFeats, totalWeights, nil
}

// buildFeatures pulls out support level from features json in device_integrations based on integration_features passed in.
// Will include entry for all feats if limit is 0, otherwise cuts off first {limit} features
func buildFeatures(featuresJSON null.JSON, feats models.IntegrationFeatureSlice) []*p_grpc.Feature {
	gfs := make([]*p_grpc.Feature, len(feats))
	if featuresJSON.IsZero() {
		return nil
	}
	for i, feat := range feats {
		supportLevel := gjson.GetBytes(featuresJSON.JSON, fmt.Sprintf(`#(featureKey=="%s").supportLevel`, feat.FeatureKey))
		slInt := int32(0)
		if supportLevel.Exists() {
			slInt = int32(supportLevel.Int())
		}

		gfs[i] = &p_grpc.Feature{
			Key:          feat.FeatureKey,
			SupportLevel: slInt,
			CssIcon:      feat.CSSIcon.String,
			DisplayName:  feat.DisplayName,
		}
	}
	return gfs
}

// calculateCompatibilityLevel calculates whether devices is bronze silver gold etc based on standard math
// currently only supports if the supportLevel is == 2
func calculateCompatibilityLevel(gfs []*p_grpc.Feature, feats models.IntegrationFeatureSlice, weights float64) CompatibilityLevel {
	if gfs == nil {
		return NoDataLevel
	}
	featureWeight := 0.0
	for _, gf := range gfs {
		// match the feature to get the FeatureWeight
		for _, feat := range feats {
			if feat.FeatureKey == gf.Key && gf.SupportLevel == 2 {
				featureWeight += feat.FeatureWeight.Float64
				break
			}
		}
	}

	return calculateMathForLevel(featureWeight, weights)
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
