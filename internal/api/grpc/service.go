package grpc

import (
	"context"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/features/device_definition/queries"
	p_grpc "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
)

type GrpcService struct {
	p_grpc.UnimplementedDeviceDefinitionServiceServer
	Mediator mediator.Mediator
}

func NewGrpcService(mediator mediator.Mediator) p_grpc.DeviceDefinitionServiceServer {
	return &GrpcService{Mediator: mediator}
}

func (s *GrpcService) GetDeviceDefinitionById(ctx context.Context, in *p_grpc.GetDeviceDefinitionRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetByIdQuery{
		DeviceDefinitionID: in.Id,
	})

	result := qryResult.(*queries.GetByIdQueryResult)

	return &p_grpc.GetDeviceDefinitionResponse{
		DeviceDefinitionId: result.DeviceDefinitionID,
		Model:              result.Model,
		Year:               int32(result.Year),
	}, nil
}
