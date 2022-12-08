package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
)

type GrpcVINDecoderService struct {
	p_grpc.VINDecoderServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcVINDecoderService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.VINDecoderServiceServer {
	return &GrpcVINDecoderService{Mediator: mediator, logger: logger}
}

func (s *GrpcReviewsService) DecodeVIN(ctx context.Context, in *p_grpc.DecodeVINRequest) (*p_grpc.DecodeVINResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.DecodeVINQuery{
		VIN: in.Vin,
	})

	result := qryResult.(*p_grpc.DecodeVINResponse)

	return result, nil
}
