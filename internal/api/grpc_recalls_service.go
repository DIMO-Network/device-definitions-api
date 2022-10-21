package api

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
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

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllIntegrationQuery{})

	integrations := qryResult.([]models.GetIntegrationQueryResult)
	result := &p_grpc.GetIntegrationResponse{}

	for _, item := range integrations {
		result.Integrations = append(result.Integrations, &p_grpc.Integration{
			Id:                      item.ID,
			Type:                    item.Type,
			Style:                   item.Style,
			Vendor:                  item.Vendor,
			AutoPiDefaultTemplateId: int32(item.AutoPiDefaultTemplateID),
			RefreshLimitSecs:        int32(item.RefreshLimitSecs),
			AutoPiPowertrainTemplate: &p_grpc.Integration_AutoPiPowertrainTemplate{
				BEV:  int32(item.AutoPiPowertrainToTemplateID[models.BEV]),
				HEV:  int32(item.AutoPiPowertrainToTemplateID[models.HEV]),
				ICE:  int32(item.AutoPiPowertrainToTemplateID[models.ICE]),
				PHEV: int32(item.AutoPiPowertrainToTemplateID[models.PHEV]),
			},
		})
	}

	return result, nil
}

func (s *GrpcService) GetRecallsByModel(ctx context.Context, in *p_grpc.GetRecallsByModelRequest) (*p_grpc.GetRecallsResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByIDQuery{})

	item := qryResult.(models.GetIntegrationQueryResult)
	result := &p_grpc.Integration{
		Id:                      item.ID,
		Type:                    item.Type,
		Style:                   item.Style,
		Vendor:                  item.Vendor,
		AutoPiDefaultTemplateId: int32(item.AutoPiDefaultTemplateID),
		RefreshLimitSecs:        int32(item.RefreshLimitSecs),
		AutoPiPowertrainTemplate: &p_grpc.Integration_AutoPiPowertrainTemplate{
			BEV:  int32(item.AutoPiPowertrainToTemplateID[models.BEV]),
			HEV:  int32(item.AutoPiPowertrainToTemplateID[models.HEV]),
			ICE:  int32(item.AutoPiPowertrainToTemplateID[models.ICE]),
			PHEV: int32(item.AutoPiPowertrainToTemplateID[models.PHEV]),
		},
	}

	return result, nil
}
