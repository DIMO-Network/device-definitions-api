package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
)

type GrpcRecallsService struct {
	p_grpc.UnimplementedRecallsServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcRecallsService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.RecallsServiceServer {
	return &GrpcRecallsService{Mediator: mediator, logger: logger}
}

func (s *GrpcService) GetRecallsByMake(ctx context.Context, in *p_grpc.GetRecallsByMakeRequest) (*p_grpc.GetRecallsResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetRecallsByMakeQuery{
		MakeID: in.MakeId,
	})

	result := qryResult.(*p_grpc.GetRecallsResponse)

	return result, nil
}

func (s *GrpcService) GetRecallsByModel(ctx context.Context, in *p_grpc.GetRecallsByModelRequest) (*p_grpc.GetRecallsResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetRecallsByModelQuery{
		DeviceDefinitionID: in.DeviceDefinitionId,
	})

	result := qryResult.(*p_grpc.GetRecallsResponse)

	return result, nil
}
