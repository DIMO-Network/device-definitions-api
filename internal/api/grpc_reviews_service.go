package api

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
)

type GrpcReviewsService struct {
	p_grpc.ReviewsServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcReviewsService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.ReviewsServiceServer {
	return &GrpcReviewsService{Mediator: mediator, logger: logger}
}

func (s *GrpcReviewsService) GetReviewsByDeviceDefinitionID(ctx context.Context, in *p_grpc.GetReviewsByDeviceDefinitionRequest) (*p_grpc.GetReviewsResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetReviewsByDeviceDefinitionIDQuery{
		DeviceDefinitionID: in.DeviceDefinitionId,
	})

	result := qryResult.(*p_grpc.GetReviewsResponse)

	return result, nil
}

func (s *GrpcReviewsService) GetReviewByID(ctx context.Context, in *p_grpc.GetReviewRequest) (*p_grpc.DeviceReview, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetReviewsByIDQuery{
		ReviewID: in.Id,
	})

	result := qryResult.(*p_grpc.DeviceReview)

	return result, nil
}

func (s *GrpcReviewsService) CreateReview(ctx context.Context, in *p_grpc.CreateReviewRequest) (*p_grpc.ReviewBaseResponse, error) {
	command := &commands.CreateReviewCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		URL:                in.Url,
		ImageURL:           in.ImageURL,
		Channel:            in.Channel,
		Comments:           in.Comments,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.CreateReviewCommandResult)

	return &p_grpc.ReviewBaseResponse{Id: result.ID}, nil
}

func (s *GrpcReviewsService) UpdateReview(ctx context.Context, in *p_grpc.UpdateReviewRequest) (*p_grpc.ReviewBaseResponse, error) {
	command := &commands.UpdateReviewCommand{
		ReviewID: in.Id,
		URL:      in.Url,
		ImageURL: in.ImageURL,
		Channel:  in.Channel,
		Comments: in.Comments,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateReviewCommandResult)

	return &p_grpc.ReviewBaseResponse{Id: result.ID}, nil
}

func (s *GrpcReviewsService) DeleteReview(ctx context.Context, in *p_grpc.DeleteReviewRequest) (*p_grpc.ReviewBaseResponse, error) {
	command := &commands.DeleteReviewCommand{
		ReviewID: in.Id,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.DeleteReviewCommandResult)

	return &p_grpc.ReviewBaseResponse{Id: result.ID}, nil
}

func (s *GrpcReviewsService) ApproveReview(ctx context.Context, in *p_grpc.ApproveReviewRequest) (*p_grpc.ReviewBaseResponse, error) {
	command := &commands.ApproveReviewCommand{
		ReviewID:   in.Id,
		ApprovedBy: in.ApprovedBy,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.ApproveReviewCommandResult)

	return &p_grpc.ReviewBaseResponse{Id: result.ID}, nil
}
