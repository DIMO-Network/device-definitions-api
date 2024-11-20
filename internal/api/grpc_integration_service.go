package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/rs/zerolog"
)

type GrpcIntegrationService struct {
	p_grpc.IntegrationServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcIntegrationService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.IntegrationServiceServer {
	return &GrpcIntegrationService{Mediator: mediator, logger: logger}
}

func (s *GrpcIntegrationService) GetIntegrationOptions(ctx context.Context, in *p_grpc.GetIntegrationOptionsRequest) (*p_grpc.GetIntegrationOptionsResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetIntegrationOptionsQuery{MakeID: in.MakeId})

	return qryResult.(*p_grpc.GetIntegrationOptionsResponse), nil
}
