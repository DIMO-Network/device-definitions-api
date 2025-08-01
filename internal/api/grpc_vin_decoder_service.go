package api

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/rs/zerolog"
)

type GrpcVinDecoderService struct {
	p_grpc.VinDecoderServiceServer
	logger           *zerolog.Logger
	decodeVINHandler *queries.DecodeVINQueryHandler
	upsertVINHandler *queries.UpsertDecodingQueryHandler
}

func NewGrpcVinDecoderService(logger *zerolog.Logger, decodeVINHandler *queries.DecodeVINQueryHandler,
	upsertVINHandler *queries.UpsertDecodingQueryHandler) p_grpc.VinDecoderServiceServer {
	return &GrpcVinDecoderService{logger: logger, decodeVINHandler: decodeVINHandler, upsertVINHandler: upsertVINHandler}
}

func (s *GrpcVinDecoderService) DecodeVin(ctx context.Context, in *p_grpc.DecodeVinRequest) (*p_grpc.DecodeVinResponse, error) {
	qry := queries.DecodeVINQuery{
		VIN:        in.Vin,
		KnownModel: in.KnownModel,
		KnownYear:  in.KnownYear,
		Country:    in.Country,
	}
	result, err := s.decodeVINHandler.Handle(ctx, &qry)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *GrpcVinDecoderService) UpsertDecoding(ctx context.Context, in *p_grpc.UpsertDecodingRequest) (*emptypb.Empty, error) {
	_, err := s.upsertVINHandler.Handle(ctx, &queries.UpsertDecodingQuery{
		VIN:          in.Vin,
		DefinitionID: in.TargetDefinitionId,
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
