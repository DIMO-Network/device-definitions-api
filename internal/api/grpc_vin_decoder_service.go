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
	qryResult, err := s.decodeVINHandler.Handle(ctx, &qry) // todo change Handle to require the actual type not mediator message
	// todo change handler to return p_grpc.DecodeVinResponse? But would need to change other places too
	if err != nil {
		return nil, err
	}

	result := qryResult.(*p_grpc.DecodeVinResponse)

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
