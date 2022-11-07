package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
)

type GrpcReviewsService struct {
	p_grpc.UnimplementedReviewsServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcReviewsService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.ReviewsServiceServer {
	return &GrpcReviewsService{Mediator: mediator, logger: logger}
}

func (s *GrpcReviewsService) GetReviewsByModel(ctx context.Context, in *p_grpc.GetReviewsByModelRequest) (*p_grpc.GetReviewsResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetReviewsByModelQuery{
		DeviceDefinitionID: in.DeviceDefinitionId,
	})

	result := qryResult.(*p_grpc.GetReviewsResponse)

	return result, nil
}
