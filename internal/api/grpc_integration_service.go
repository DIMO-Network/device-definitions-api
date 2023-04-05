package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
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

func (s *GrpcIntegrationService) GetCompatibilitiesByMake(ctx context.Context, in *p_grpc.GetCompatibilitiesByMakeRequest) (*p_grpc.GetCompatibilitiesByMakeResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetCompatibilitiesByMakeQuery{
		MakeID:        in.MakeId,
		IntegrationID: in.IntegrationId,
		Region:        in.Region,
		Skip:          in.Skip,
		Take:          in.Take,
	})

	result := qryResult.(*p_grpc.GetCompatibilitiesByMakeResponse)
	return result, nil
}

func (s *GrpcIntegrationService) GetIntegrationFeatures(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetIntegrationFeatureListResponse, error) {
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

func (s *GrpcIntegrationService) GetCompatibilityByDeviceArray(ctx context.Context, in *p_grpc.GetCompatibilityByDeviceArrayRequest) (*p_grpc.GetCompatibilityByDeviceArrayResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetCompatibilityByDeviceDefinitionArrayQuery{DeviceDefinitionID: in.DeviceDefinitionIds})
	return qryResult.(*p_grpc.GetCompatibilityByDeviceArrayResponse), nil
}

func (s *GrpcIntegrationService) GetIntegrationOptions(ctx context.Context, in *p_grpc.GetIntegrationOptionsRequest) (*p_grpc.GetIntegrationOptionsResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetIntegrationOptionsQuery{MakeID: in.MakeId})

	return qryResult.(*p_grpc.GetIntegrationOptionsResponse), nil
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
