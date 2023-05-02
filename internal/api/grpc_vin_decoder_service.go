package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
)

type GrpcVinDecoderService struct {
	p_grpc.VinDecoderServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcVinDecoderService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.VinDecoderServiceServer {
	return &GrpcVinDecoderService{Mediator: mediator, logger: logger}
}

func (s *GrpcVinDecoderService) DecodeVin(ctx context.Context, in *p_grpc.DecodeVinRequest) (*p_grpc.DecodeVinResponse, error) {
	qryResult, err := s.Mediator.Send(ctx, &queries.DecodeVINQuery{
		VIN:        in.Vin,
		KnownModel: in.KnownModel,
		KnownYear:  in.KnownYear,
		Country:    in.Country,
	})
	if err != nil {
		return nil, err
	}

	result := qryResult.(*p_grpc.DecodeVinResponse)

	return result, nil
}
