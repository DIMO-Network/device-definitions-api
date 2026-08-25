package api

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/rs/zerolog"
)

type GrpcDefinitionsService struct {
	p_grpc.DeviceDefinitionServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
	dbs      *db.ReaderWriter
}

func NewGrpcService(mediator mediator.Mediator, logger *zerolog.Logger, dbs func() *db.ReaderWriter) p_grpc.DeviceDefinitionServiceServer {
	return &GrpcDefinitionsService{Mediator: mediator, logger: logger, dbs: dbs()}
}

//** Device Definitions
// Definition create/update/list moved to the definitions-worker (R2 catalog)
// and the public identity-api / catalog endpoints. The RPCs remain registered
// for wire compatibility but are no longer implemented.

func (s *GrpcDefinitionsService) GetFilteredDeviceDefinition(_ context.Context, _ *p_grpc.FilterDeviceDefinitionRequest) (*p_grpc.GetFilteredDeviceDefinitionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "device definition reads moved to identity-api and the definitions catalog")
}

func (s *GrpcDefinitionsService) CreateDeviceDefinition(_ context.Context, _ *p_grpc.CreateDeviceDefinitionRequest) (*p_grpc.CreateDeviceDefinitionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "device definition writes moved to the definitions-worker")
}

func (s *GrpcDefinitionsService) UpdateDeviceDefinition(_ context.Context, _ *p_grpc.UpdateDeviceDefinitionRequest) (*p_grpc.BaseResponse, error) {
	return nil, status.Error(codes.Unimplemented, "device definition writes moved to the definitions-worker")
}

//** Integrations

func (s *GrpcDefinitionsService) GetIntegrations(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetIntegrationResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllIntegrationQuery{})

	integrations := qryResult.([]coremodels.GetIntegrationQueryResult)
	result := &p_grpc.GetIntegrationResponse{}

	for _, item := range integrations {
		intg := &p_grpc.Integration{
			Id:                      item.ID,
			Type:                    item.Type,
			Style:                   item.Style,
			Vendor:                  item.Vendor,
			AutoPiDefaultTemplateId: int32(item.AutoPiDefaultTemplateID),
			RefreshLimitSecs:        int32(item.RefreshLimitSecs),
			TokenId:                 uint64(item.TokenID),
			AutoPiPowertrainTemplate: &p_grpc.Integration_AutoPiPowertrainTemplate{
				BEV:  int32(item.AutoPiPowertrainToTemplateID[coremodels.BEV]),
				HEV:  int32(item.AutoPiPowertrainToTemplateID[coremodels.HEV]),
				ICE:  int32(item.AutoPiPowertrainToTemplateID[coremodels.ICE]),
				PHEV: int32(item.AutoPiPowertrainToTemplateID[coremodels.PHEV]),
			},
			Points:              int64(item.Points),
			ManufacturerTokenId: uint64(item.ManufacturerTokenID),
		}

		result.Integrations = append(result.Integrations, intg)
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetIntegrationByID(ctx context.Context, in *p_grpc.GetIntegrationRequest) (*p_grpc.Integration, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetIntegrationByIDQuery{IntegrationID: in.Id})

	item := qryResult.(coremodels.GetIntegrationQueryResult)
	return s.prepareIntegrationResponse(item)
}

func (s *GrpcDefinitionsService) GetIntegrationByTokenID(ctx context.Context, in *p_grpc.GetIntegrationByTokenIDRequest) (*p_grpc.Integration, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetIntegrationByTokenIDQuery{
		TokenID: int(in.TokenId),
	})

	item := qryResult.(coremodels.GetIntegrationQueryResult)
	return s.prepareIntegrationResponse(item)
}

//** Device Styles / Trims

func (s *GrpcDefinitionsService) CreateDeviceStyle(ctx context.Context, in *p_grpc.CreateDeviceStyleRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceStyleCommand{
		DefinitionID:    in.DefinitionId,
		Name:            in.Name,
		ExternalStyleID: in.ExternalStyleId,
		Source:          in.Source,
		SubModel:        in.SubModel,
	})

	result := commandResult.(commands.CreateDeviceStyleCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) GetDeviceStyleByID(ctx context.Context, in *p_grpc.GetDeviceStyleByIDRequest) (*p_grpc.DeviceStyle, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByIDQuery{
		DeviceStyleID: in.Id,
	})

	ds := qryResult.(coremodels.GetDeviceStyleQueryResult)
	result := &p_grpc.DeviceStyle{
		Id:              ds.ID,
		Source:          ds.Source,
		SubModel:        ds.SubModel,
		Name:            ds.Name,
		ExternalStyleId: ds.ExternalStyleID,
		DefinitionId:    ds.DefinitionID,
	}

	if len(ds.DeviceDefinition.DeviceAttributes) > 0 {
		for _, prop := range ds.DeviceDefinition.DeviceAttributes {
			result.DeviceAttributes = append(result.DeviceAttributes, &p_grpc.DeviceTypeAttribute{
				Name:        prop.Name,
				Label:       prop.Label,
				Value:       prop.Value,
				Description: prop.Description,
				Required:    prop.Required,
				Options:     prop.Option,
			})
		}
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceStylesByDeviceDefinitionID(ctx context.Context, in *p_grpc.GetDeviceStyleByDeviceDefinitionIDRequest) (*p_grpc.GetDeviceStyleResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByDeviceDefinitionIDQuery{
		DefinitionID: in.Id,
	})

	styles := qryResult.([]coremodels.GetDeviceStyleQueryResult)

	result := &p_grpc.GetDeviceStyleResponse{}

	for _, ds := range styles {
		result.DeviceStyles = append(result.DeviceStyles, &p_grpc.DeviceStyle{
			Id:              ds.ID,
			Source:          ds.Source,
			SubModel:        ds.SubModel,
			Name:            ds.Name,
			ExternalStyleId: ds.ExternalStyleID,
			DefinitionId:    ds.DefinitionID,
		})
	}

	return result, nil
}

func (s *GrpcDefinitionsService) UpdateDeviceStyle(ctx context.Context, in *p_grpc.UpdateDeviceStyleRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.UpdateDeviceStyleCommand{
		ID:              in.Id,
		Name:            in.Name,
		ExternalStyleID: in.ExternalStyleId,
		DefinitionID:    in.DefinitionId,
		Source:          in.Source,
		SubModel:        in.SubModel,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceStyleCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

//** Device Types / Attributes

func (s *GrpcDefinitionsService) GetDeviceTypesByID(ctx context.Context, in *p_grpc.GetDeviceTypeByIDRequest) (*p_grpc.GetDeviceTypeResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceTypeByIDQuery{
		DeviceTypeID: in.Id,
	})

	dt := qryResult.(coremodels.GetDeviceTypeQueryResult)
	result := &p_grpc.GetDeviceTypeResponse{
		Id:   dt.ID,
		Name: dt.Name,
	}

	for _, prop := range dt.Attributes {
		result.Attributes = append(result.Attributes, &p_grpc.DeviceTypeAttribute{
			Name:         prop.Name,
			Label:        prop.Label,
			Description:  prop.Description,
			Required:     prop.Required,
			DefaultValue: prop.DefaultValue,
			Options:      prop.Options,
		})
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceTypes(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDeviceTypeListResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceTypeQuery{})

	dt := qryResult.([]coremodels.GetDeviceTypeQueryResult)

	items := make([]*p_grpc.GetDeviceTypeResponse, len(dt))
	for i, v := range dt {
		items[i] = &p_grpc.GetDeviceTypeResponse{
			Id:   v.ID,
			Name: v.Name,
		}

		items[i].Attributes = make([]*p_grpc.DeviceTypeAttribute, len(v.Attributes))
		for x, attr := range v.Attributes {
			items[i].Attributes[x] = &p_grpc.DeviceTypeAttribute{
				Name:         attr.Name,
				Type:         attr.Type,
				Description:  attr.Description,
				Required:     attr.Required,
				DefaultValue: attr.DefaultValue,
				Options:      attr.Options,
			}
		}

	}

	result := &p_grpc.GetDeviceTypeListResponse{DeviceTypes: items}

	return result, nil
}

func (s *GrpcDefinitionsService) CreateDeviceType(ctx context.Context, in *p_grpc.CreateDeviceTypeRequest) (*p_grpc.BaseResponse, error) {
	command := &commands.CreateDeviceTypeCommand{
		ID:   in.Id,
		Name: in.Name,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.CreateDeviceTypeCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) UpdateDeviceType(ctx context.Context, in *p_grpc.UpdateDeviceTypeRequest) (*p_grpc.BaseResponse, error) {
	command := &commands.UpdateDeviceTypeCommand{
		ID:   in.Id,
		Name: in.Name,
	}

	if len(in.Attributes) > 0 {
		for _, attr := range in.Attributes {
			command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.CreateDeviceTypeAttribute{
				Name:         attr.Name,
				Type:         attr.Type,
				Label:        attr.Label,
				Description:  attr.Description,
				Options:      attr.Options,
				Required:     attr.Required,
				DefaultValue: attr.DefaultValue,
			})
		}
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceTypeCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) DeleteDeviceType(ctx context.Context, in *p_grpc.DeleteDeviceTypeRequest) (*p_grpc.BaseResponse, error) {
	command := &commands.DeleteDeviceTypeCommand{
		ID: in.Id,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.DeleteDeviceTypeCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) prepareIntegrationResponse(integration coremodels.GetIntegrationQueryResult) (*p_grpc.Integration, error) {
	return &p_grpc.Integration{
		Id:                      integration.ID,
		Type:                    integration.Type,
		Style:                   integration.Style,
		Vendor:                  integration.Vendor,
		AutoPiDefaultTemplateId: int32(integration.AutoPiDefaultTemplateID),
		RefreshLimitSecs:        int32(integration.RefreshLimitSecs),
		TokenId:                 uint64(integration.TokenID),
		AutoPiPowertrainTemplate: &p_grpc.Integration_AutoPiPowertrainTemplate{
			BEV:  int32(integration.AutoPiPowertrainToTemplateID[coremodels.BEV]),
			HEV:  int32(integration.AutoPiPowertrainToTemplateID[coremodels.HEV]),
			ICE:  int32(integration.AutoPiPowertrainToTemplateID[coremodels.ICE]),
			PHEV: int32(integration.AutoPiPowertrainToTemplateID[coremodels.PHEV]),
		},
		Points:              int64(integration.Points),
		ManufacturerTokenId: uint64(integration.ManufacturerTokenID),
	}, nil
}
