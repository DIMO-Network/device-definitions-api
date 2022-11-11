package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	elasticModels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch/models"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcIntegrationService struct {
	p_grpc.IntegrationServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcIntegrationService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.IntegrationServiceServer {
	return &GrpcIntegrationService{Mediator: mediator, logger: logger}
}

func (s *GrpcIntegrationService) GetIntegrationFeatureByID(ctx context.Context, in *p_grpc.GetIntegrationFeatureByIDRequest) (*p_grpc.GetIntegrationFeatureResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetIntegrationFeatureByIDQuery{
		ID: in.Id,
	})

	feature := qryResult.(models.GetIntegrationFeatureQueryResult)
	result := &p_grpc.GetIntegrationFeatureResponse{
		FeatureKey:      feature.FeatureKey,
		CssIcon:         feature.CSSIcon,
		DisplayName:     feature.DisplayName,
		ElasticProperty: feature.ElasticProperty,
		FeatureWeight:   float32(feature.FeatureWeight),
	}

	return result, nil
}

func (s *GrpcIntegrationService) GetDeviceCompatibilities(ctx context.Context, in *p_grpc.GetDeviceCompatibilityListRequest) (*p_grpc.GetDeviceCompatibilityListResponse, error) {
	logger := s.logger.With().Str("rpc", "GetDeviceCompatibilities").Logger()
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceCompatibilityQuery{
		MakeID:        in.MakeId,
		IntegrationID: in.IntegrationId,
		Region:        in.Region,
		Cursor:        in.Cursor,
		Size:          in.Size,
	})

	deviceCompatibilities := qryResult.(queries.GetDeviceCompatibilityQueryResult)

	result := &p_grpc.GetDeviceCompatibilityListResponse{}

	integFeats := deviceCompatibilities.IntegrationFeatures
	totalWeightsCount := 0.0
	for _, v := range deviceCompatibilities.IntegrationFeatures {
		totalWeightsCount += v.FeatureWeight
	}
	dcMap := make(map[string][]*p_grpc.DeviceCompatibilities)

	cursor := ""
	if len(deviceCompatibilities.DeviceDefinitions) > 0 {
		cursor = deviceCompatibilities.DeviceDefinitions[len(deviceCompatibilities.DeviceDefinitions)-1].ID
	}
	// Group by model name.
	for _, v := range deviceCompatibilities.DeviceDefinitions {
		if len(v.R.DeviceIntegrations) == 0 {
			// This should never happen, because of the inner join.
			logger.Error().Msg("No integrations for this definition.")
			continue
		}

		di := v.R.DeviceIntegrations[0]

		if di.Features.IsZero() {
			// This should never happen, because we filtered for "features IS NOT NULL".
			logger.Error().Msg("Feature column was null.")
			continue
		}
		res := &p_grpc.DeviceCompatibilities{Year: int32(v.Year)}

		var feats []*p_grpc.Feature
		var features []elasticModels.DeviceIntegrationFeatures

		err := di.Features.Unmarshal(&features)
		if err != nil {
			logger.Err(err).Msg("Error unmarshaling features JSON blob.")
			continue
		}

		ifeat := map[string]queries.FeatureDetails{}

		for _, f := range features {
			ft := &p_grpc.Feature{
				Key:          integFeats[f.FeatureKey].DisplayName,
				CssIcon:      integFeats[f.FeatureKey].CSSIcon,
				SupportLevel: int32(f.SupportLevel),
			}

			fts := queries.FeatureDetails{
				FeatureWeight: integFeats[f.FeatureKey].FeatureWeight,
				SupportLevel:  int32(f.SupportLevel),
			}

			ifeat[f.FeatureKey] = fts

			feats = append(feats, ft)
		}

		level := queries.GetDeviceCompatibilityLevel(ifeat, totalWeightsCount)
		res.Features = feats
		res.Level = level
		dcMap[v.Model] = append(dcMap[v.Model], res)
	}

	for k, v := range dcMap {
		dcr := &p_grpc.DeviceCompatibilityList{Name: k, Years: v}
		result.Models = append(result.Models, dcr)
		result.Cursor = cursor
	}

	return result, nil
}

func (s *GrpcIntegrationService) GetIntegrationFeatures(ctx context.Context, in *emptypb.Empty) (*p_grpc.GetIntegrationFeatureListResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllIntegrationFeatureQuery{})

	dt := qryResult.([]models.GetIntegrationFeatureQueryResult)

	items := make([]*p_grpc.GetIntegrationFeatureResponse, len(dt))
	for i, feature := range dt {
		items[i] = &p_grpc.GetIntegrationFeatureResponse{
			FeatureKey:      feature.FeatureKey,
			CssIcon:         feature.CSSIcon,
			DisplayName:     feature.DisplayName,
			ElasticProperty: feature.ElasticProperty,
			FeatureWeight:   float32(feature.FeatureWeight),
		}
	}

	result := &p_grpc.GetIntegrationFeatureListResponse{IntegrationFeatures: items}

	return result, nil
}

func (s *GrpcIntegrationService) GetCompatibilityByDeviceDefinition(ctx context.Context, in *p_grpc.GetCompatibilityByDeviceDefinitionRequest) (*p_grpc.GetDeviceCompatibilitiesResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetCompatibilityByDeviceDefinitionQuery{DeviceDefinitionID: in.DeviceDefinitionId})

	return qryResult.(*p_grpc.GetDeviceCompatibilitiesResponse), nil
}

func (s *GrpcIntegrationService) CreateIntegrationFeature(ctx context.Context, in *p_grpc.CreateOrUpdateIntegrationFeatureRequest) (*p_grpc.IntegrationBaseResponse, error) {
	command := &commands.CreateIntegrationFeatureCommand{
		ID:              in.Id,
		CSSIcon:         in.CssIcon,
		DisplayName:     in.DisplayName,
		ElasticProperty: in.ElasticProperty,
		FeatureWeight:   float64(in.FeatureWeight),
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.CreateIntegrationFeatureCommandResult)

	return &p_grpc.IntegrationBaseResponse{Id: result.ID}, nil
}

func (s *GrpcIntegrationService) UpdateIntegrationFeature(ctx context.Context, in *p_grpc.CreateOrUpdateIntegrationFeatureRequest) (*p_grpc.IntegrationBaseResponse, error) {
	command := &commands.UpdateIntegrationFeatureCommand{
		ID:              in.Id,
		CSSIcon:         in.CssIcon,
		DisplayName:     in.DisplayName,
		ElasticProperty: in.ElasticProperty,
		FeatureWeight:   float64(in.FeatureWeight),
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateIntegrationFeatureResult)

	return &p_grpc.IntegrationBaseResponse{Id: result.ID}, nil
}

func (s *GrpcIntegrationService) DeleteIntegrationFeature(ctx context.Context, in *p_grpc.DeleteIntegrationFeatureRequest) (*p_grpc.IntegrationBaseResponse, error) {
	command := &commands.DeleteIntegrationFeatureCommand{
		ID: in.Id,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.DeleteIntegrationFeatureCommand)

	return &p_grpc.IntegrationBaseResponse{Id: result.ID}, nil
}
