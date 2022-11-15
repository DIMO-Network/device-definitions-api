package queries

import (
	"context"
	"fmt"
	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/tidwall/gjson"
	"github.com/volatiletech/null/v8"
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
	// could pull this one from cache since doesn't change often
	integFeats, err := models.IntegrationFeatures(qm.Limit(100)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: errors.Wrap(err, "failed to get integration_features"),
		}
	}
	totalWeights := 0.0
	for _, v := range integFeats {
		if !v.FeatureWeight.IsZero() {
			totalWeights += v.FeatureWeight.Float64
		}
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

// buildFeatures pulls out support level from features json in device_integrations based on integration_features passed in. Will include entry for all feats
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

func calculateCompatibilityLevel(gfs []*p_grpc.Feature, feats models.IntegrationFeatureSlice, weights float64) CompatibilityLevel {
	if gfs == nil {
		return NoDataLevel
	}
	featureWeight := 0.0
	for _, gf := range gfs {
		// match the feature to get the FeatureWeight
		for _, feat := range feats {
			if feat.FeatureKey == gf.Key && gf.SupportLevel > 0 {
				featureWeight += feat.FeatureWeight.Float64
				break
			}
		}
	}

	return calculateMathForLevel(featureWeight, weights)
}
