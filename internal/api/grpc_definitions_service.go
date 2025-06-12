package api

import (
	"context"
	"encoding/json"

	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/friendsofgo/errors"
	"github.com/rs/zerolog"
)

type GrpcDefinitionsService struct {
	p_grpc.DeviceDefinitionServiceServer
	Mediator          mediator.Mediator
	logger            *zerolog.Logger
	dbs               *db.ReaderWriter
	onChainDeviceDefs gateways.DeviceDefinitionOnChainService
	queryInstance     *contracts.Registry
	identity          gateways.IdentityAPI
}

func NewGrpcService(mediator mediator.Mediator, logger *zerolog.Logger, dbs func() *db.ReaderWriter,
	onChainDefs gateways.DeviceDefinitionOnChainService, queryInstance *contracts.Registry, identity gateways.IdentityAPI) p_grpc.DeviceDefinitionServiceServer {
	return &GrpcDefinitionsService{Mediator: mediator, logger: logger, dbs: dbs(), onChainDeviceDefs: onChainDefs, queryInstance: queryInstance, identity: identity}
}

//** Device Definitions

// GetFilteredDeviceDefinition used by: admin, cs-support-platform
func (s *GrpcDefinitionsService) GetFilteredDeviceDefinition(ctx context.Context, in *p_grpc.FilterDeviceDefinitionRequest) (*p_grpc.GetFilteredDeviceDefinitionsResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByDynamicFilterQuery{
		DefinitionID:    in.DefinitionId,
		MakeSlug:        in.MakeSlug,
		Year:            int(in.Year),
		Model:           in.Model,
		VerifiedVinList: in.VerifiedVinList,
		PageIndex:       int(in.PageIndex),
		PageSize:        int(in.PageSize),
	})

	ddResult := qryResult.([]queries.DeviceDefinitionQueryResponse)

	result := &p_grpc.GetFilteredDeviceDefinitionsResponse{}

	for _, deviceDefinition := range ddResult {
		var ei map[string]string
		var extIDs []*p_grpc.ExternalID
		if err := deviceDefinition.ExternalIDs.Unmarshal(&ei); err != nil {
			for vendor, id := range ei {
				extIDs = append(extIDs, &p_grpc.ExternalID{
					Vendor: vendor,
					Id:     id,
				})
			}
		}
		result.Items = append(result.Items, &p_grpc.FilterDeviceDefinitionsReponse{
			Id:           deviceDefinition.ID,
			NameSlug:     deviceDefinition.NameSlug,
			Model:        deviceDefinition.Model,
			Year:         int32(deviceDefinition.Year),
			ImageUrl:     deviceDefinition.ImageURL.String,
			CreatedAt:    deviceDefinition.CreatedAt.UnixMilli(),
			UpdatedAt:    deviceDefinition.UpdatedAt.UnixMilli(),
			Metadata:     string(deviceDefinition.Metadata.JSON),
			Verified:     deviceDefinition.Verified,
			DeviceMakeId: deviceDefinition.DeviceMakeID,
			Make:         deviceDefinition.Make,
			ExternalIds:  extIDs,
		})
	}

	return result, nil
}

// CreateDeviceDefinition used by: dimo-admin
func (s *GrpcDefinitionsService) CreateDeviceDefinition(ctx context.Context, in *p_grpc.CreateDeviceDefinitionRequest) (*p_grpc.CreateDeviceDefinitionResponse, error) {
	// todo we could call the command directly, but maintain testability, what about the update, could it share logic
	command := &commands.CreateDeviceDefinitionCommand{
		Source:             in.Source,
		Make:               in.Make,
		Model:              in.Model,
		Year:               int(in.Year),
		DeviceTypeID:       in.DeviceTypeId,
		HardwareTemplateID: in.HardwareTemplateId,
		Verified:           in.Verified,
	}

	if len(in.DeviceAttributes) > 0 {
		for _, attribute := range in.DeviceAttributes {
			command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
				Name:  attribute.Name,
				Value: attribute.Value,
			})
		}
	}

	commandResult, _ := s.Mediator.Send(ctx, command)
	result := commandResult.(commands.CreateDeviceDefinitionCommandResult)

	return &p_grpc.CreateDeviceDefinitionResponse{Id: result.ID, NameSlug: result.NameSlug}, nil
}

// UpdateDeviceDefinition is used by admin tool to update tableland properties of a dd, and a couple augmented properties
func (s *GrpcDefinitionsService) UpdateDeviceDefinition(ctx context.Context, in *p_grpc.UpdateDeviceDefinitionRequest) (*p_grpc.BaseResponse, error) {
	// if verified = true, send update request to tableland
	dm, err := s.identity.GetManufacturer(stringutils.SlugString(in.ManufacturerName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to find device make")
	}
	// get manufacturer from chain
	manufacturerID, err := s.queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, dm.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to GetManufacturerIdByName for update: %s", dm.Name)
	}
	ddTbland, err := s.onChainDeviceDefs.GetDefinitionTableland(ctx, manufacturerID, in.DeviceDefinitionId) // repurposed for definitionID
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find device definition in tableland for update: %s", in.DeviceDefinitionId)
	}
	shouldUpdate := false
	metadata := gateways.BuildDeviceTypeAttributesTbland(in.DeviceAttributes)
	tblMetadata, err := json.Marshal(ddTbland.Metadata)
	if err != nil {
		s.logger.Err(err).Msgf("failed to unmarshall metadata for: %s", in.DeviceDefinitionId)
	}
	// check for any changes
	if string(tblMetadata) != metadata || ddTbland.DeviceType != in.DeviceTypeId || ddTbland.ImageURI != in.ImageUrl {
		shouldUpdate = true
	}
	if shouldUpdate {
		// on chain portion of update
		_, err = s.onChainDeviceDefs.Update(ctx, dm.Name, contracts.DeviceDefinitionUpdateInput{
			Id:         in.DeviceDefinitionId, // name slug
			Metadata:   metadata,
			Ksuid:      ddTbland.KSUID,
			DeviceType: in.DeviceTypeId,
			ImageURI:   in.ImageUrl,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to update device definition on chain")
		}
	}

	return &p_grpc.BaseResponse{Id: in.DeviceDefinitionId}, nil
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
