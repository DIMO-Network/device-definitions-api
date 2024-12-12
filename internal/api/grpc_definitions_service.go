package api

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

type GrpcDefinitionsService struct {
	p_grpc.DeviceDefinitionServiceServer
	Mediator          mediator.Mediator
	logger            *zerolog.Logger
	dbs               *db.ReaderWriter
	onChainDeviceDefs gateways.DeviceDefinitionOnChainService
	queryInstance     *contracts.Registry
}

func NewGrpcService(mediator mediator.Mediator, logger *zerolog.Logger, dbs func() *db.ReaderWriter,
	onChainDefs gateways.DeviceDefinitionOnChainService, queryInstance *contracts.Registry) p_grpc.DeviceDefinitionServiceServer {
	return &GrpcDefinitionsService{Mediator: mediator, logger: logger, dbs: dbs(), onChainDeviceDefs: onChainDefs, queryInstance: queryInstance}
}

//** Device Definitions

func (s *GrpcDefinitionsService) GetDeviceDefinitionByID(ctx context.Context, in *p_grpc.GetDeviceDefinitionRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByIDsQuery{
		DeviceDefinitionID: in.Ids,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionBySlug(ctx context.Context, in *p_grpc.GetDeviceDefinitionBySlugRequest) (*p_grpc.GetDeviceDefinitionItemResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionBySlugQuery{
		DefinitionID: in.Slug,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionItemResponse)
	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionByMMY(ctx context.Context, in *p_grpc.GetDeviceDefinitionByMMYRequest) (*p_grpc.GetDeviceDefinitionItemResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByMakeModelYearQuery{
		Make:  in.Make,
		Model: in.Model,
		Year:  int(in.Year),
	})

	dd := qryResult.(*coremodels.GetDeviceDefinitionQueryResult)
	result := common.BuildFromQueryResultToGRPC(dd)

	return result, nil
}

func (s *GrpcDefinitionsService) GetFilteredDeviceDefinition(ctx context.Context, in *p_grpc.FilterDeviceDefinitionRequest) (*p_grpc.GetFilteredDeviceDefinitionsResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByDynamicFilterQuery{
		MakeID:             in.MakeId,
		IntegrationID:      in.IntegrationId,
		DeviceDefinitionID: in.DeviceDefinitionId,
		Year:               int(in.Year),
		Model:              in.Model,
		VerifiedVinList:    in.VerifiedVinList,
		PageIndex:          int(in.PageIndex),
		PageSize:           int(in.PageSize),
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
			Source:       deviceDefinition.Source.String,
			Verified:     deviceDefinition.Verified,
			External:     deviceDefinition.ExternalID.String,
			DeviceMakeId: deviceDefinition.DeviceMakeID,
			Make:         deviceDefinition.Make,
			ExternalIds:  extIDs,
		})
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionBySource(in *p_grpc.GetDeviceDefinitionBySourceRequest, stream p_grpc.DeviceDefinitionService_GetDeviceDefinitionBySourceServer) error {
	qryResult, _ := s.Mediator.Send(context.Background(), &queries.GetDeviceDefinitionBySourceQuery{
		Source: in.Source,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	for _, dd := range result.DeviceDefinitions {
		if err := stream.Send(dd); err != nil {
			return err
		}
	}

	return nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionWithoutImages(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionWithoutImageQuery{})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

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

func (s *GrpcDefinitionsService) GetDeviceDefinitionAll(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDeviceDefinitionAllResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceDefinitionGroupQuery{})

	result := &p_grpc.GetDeviceDefinitionAllResponse{}

	allDevices := qryResult.([]queries.GetAllDeviceDefinitionGroupQueryResult)

	for _, device := range allDevices {
		item := &p_grpc.GetDeviceDefinitionAllItemResponse{Make: device.Make}

		for _, model := range device.Models {
			itemModel := &p_grpc.GetDeviceDefinitionAllItemResponse_GetDeviceModels{
				Model: model.Model,
			}

			for _, modelYear := range model.Years {
				itemYear := &p_grpc.GetDeviceDefinitionAllItemResponse_GetDeviceModelYears{
					Year:               int32(modelYear.Year),
					DeviceDefinitionID: modelYear.DeviceDefinitionID,
				}
				itemModel.Years = append(itemModel.Years, itemYear)
			}

			item.Models = append(item.Models, itemModel)
		}

		result.Items = append(result.Items, item)
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitions(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceDefinitionQuery{})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcDefinitionsService) GetDevicesMMY(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDevicesMMYResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDevicesMMYQuery{})
	result := qryResult.(*p_grpc.GetDevicesMMYResponse)
	return result, nil
}

// UpdateDeviceDefinition is used by admin tool to update tableland properties of a dd, and a couple augmented properties
func (s *GrpcDefinitionsService) UpdateDeviceDefinition(ctx context.Context, in *p_grpc.UpdateDeviceDefinitionRequest) (*p_grpc.BaseResponse, error) {
	// if verified = true, send update request to tableland

	dbDD, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.NameSlug.EQ(in.DeviceDefinitionId)).One(ctx, s.dbs.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find device definition")
	}
	dm, err := models.FindDeviceMake(ctx, s.dbs.Reader, dbDD.DeviceMakeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find device make")
	}
	// on chain portion of update
	trxHash := new(string)
	if dbDD.Verified {
		// check for any changes
		manufacturerID, err := s.queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, dm.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to GetManufacturerIdByName for update: %s", dm.Name)
		}
		ddTbland, err := s.onChainDeviceDefs.GetDefinitionTableland(ctx, manufacturerID, in.DeviceDefinitionId)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find device definition in tableland for update: %s", in.DeviceDefinitionId)
		}
		shouldUpdate := false
		metadata := gateways.BuildDeviceTypeAttributesTbland(in.DeviceAttributes)
		tblMetadata, err := json.Marshal(ddTbland.Metadata)
		if err != nil {
			s.logger.Err(err).Msgf("failed to unmarshall metadata for: %s", in.DeviceDefinitionId)
		}
		if string(tblMetadata) != metadata || ddTbland.DeviceType != in.DeviceTypeId || ddTbland.KSUID != dbDD.ID || ddTbland.ImageURI != in.ImageUrl {
			shouldUpdate = true
		}
		if shouldUpdate {
			trxHash, err = s.onChainDeviceDefs.Update(ctx, dm.Name, contracts.DeviceDefinitionUpdateInput{
				Id:         in.DeviceDefinitionId, // name slug
				Metadata:   metadata,
				Ksuid:      dbDD.ID,
				DeviceType: in.DeviceTypeId,
				ImageURI:   in.ImageUrl,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to update device definition on chain")
			}
		}
	}
	// database portion of update, note we only update properties that are not on-chain as we intend to deprecate duplicate fields
	dbDD.Verified = in.Verified
	dbDD.HardwareTemplateID = null.StringFrom(in.HardwareTemplateId)
	if trxHash != nil {
		dbDD.TRXHashHex = append(dbDD.TRXHashHex, *trxHash)
	}
	_, err = dbDD.Update(ctx, s.dbs.Writer, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "failed to update device definition")
	}
	return &p_grpc.BaseResponse{Id: dbDD.NameSlug}, nil
}

func (s *GrpcDefinitionsService) GetDeviceImagesByIDs(ctx context.Context, in *p_grpc.GetDeviceDefinitionRequest) (*p_grpc.GetDeviceImagesResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionImagesByIDsQuery{
		DeviceDefinitionID: in.Ids,
	})

	return qryResult.(*p_grpc.GetDeviceImagesResponse), nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionsWithHardwareTemplate(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDevicesMMYResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDefinitionsWithHWTemplateQuery{})

	return qryResult.(*p_grpc.GetDevicesMMYResponse), nil
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

func (s *GrpcDefinitionsService) GetDeviceDefinitionIntegration(ctx context.Context, in *p_grpc.GetDeviceDefinitionIntegrationRequest) (*p_grpc.GetDeviceDefinitionIntegrationResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionWithRelsQuery{
		DeviceDefinitionID: in.Id,
	})

	queryResult := qryResult.([]queries.GetDeviceDefinitionWithRelsQueryResult)

	result := &p_grpc.GetDeviceDefinitionIntegrationResponse{}

	for _, queryResult := range queryResult {
		result.Integrations = append(result.Integrations, &p_grpc.DeviceIntegration{
			Integration: &p_grpc.Integration{
				Id:     queryResult.ID,
				Type:   queryResult.Type,
				Style:  queryResult.Style,
				Vendor: queryResult.Vendor,
			},
			Region: queryResult.Region,
		})
	}

	return result, nil
}

func (s *GrpcDefinitionsService) CreateIntegration(ctx context.Context, in *p_grpc.CreateIntegrationRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateIntegrationCommand{
		Vendor:  in.Vendor,
		Style:   in.Style,
		Type:    in.Type,
		TokenID: int(in.TokenId),
	})

	result := commandResult.(commands.CreateIntegrationCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) CreateDeviceIntegration(ctx context.Context, in *p_grpc.CreateDeviceIntegrationRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.CreateDeviceIntegrationCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		IntegrationID:      in.IntegrationId,
		Region:             in.Region,
	}

	if len(in.Features) > 0 {
		for _, feature := range in.Features {
			command.Features = append(command.Features, &coremodels.UpdateDeviceIntegrationFeatureAttribute{
				FeatureKey:   feature.FeatureKey,
				SupportLevel: int16(feature.SupportLevel),
			})
		}
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.CreateDeviceIntegrationCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionByMakeAndYearRange(ctx context.Context, in *p_grpc.GetDeviceDefinitionByMakeAndYearRangeRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceDefinitionByMakeYearRangeQuery{
		Make:      in.Make,
		StartYear: in.StartYear,
		EndYear:   in.EndYear,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionOnChainByID(ctx context.Context, in *p_grpc.GetDeviceDefinitionRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionOnChainByIDQuery{
		DeviceDefinitionID: in.Ids[0],
		MakeSlug:           in.MakeSlug,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionsOnChain(ctx context.Context, in *p_grpc.FilterDeviceDefinitionRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceDefinitionOnChainQuery{
		DeviceDefinitionID: in.DeviceDefinitionId,
		Year:               int(in.Year),
		Model:              in.Model,
		MakeSlug:           in.MakeSlug,
		PageSize:           in.PageSize,
		PageIndex:          in.PageIndex,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceDefinitionBySlugName(ctx context.Context, in *p_grpc.GetDeviceDefinitionBySlugNameRequest) (*p_grpc.GetDeviceDefinitionItemResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionBySlugNameQuery{
		Slug: in.Slug,
	})

	dd := qryResult.(*coremodels.GetDeviceDefinitionQueryResult)
	result := common.BuildFromQueryResultToGRPC(dd)
	return result, nil
}

//** Device Styles / Trims

func (s *GrpcDefinitionsService) CreateDeviceStyle(ctx context.Context, in *p_grpc.CreateDeviceStyleRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceStyleCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		Name:               in.Name,
		ExternalStyleID:    in.ExternalStyleId,
		Source:             in.Source,
		SubModel:           in.SubModel,
		HardwareTemplateID: in.HardwareTemplateId,
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
		Id:                 ds.ID,
		Source:             ds.Source,
		SubModel:           ds.SubModel,
		Name:               ds.Name,
		ExternalStyleId:    ds.ExternalStyleID,
		DeviceDefinitionId: ds.DeviceDefinitionID,
		HardwareTemplateId: ds.HardwareTemplateID,
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
		DeviceDefinitionID: in.Id,
	})

	styles := qryResult.([]coremodels.GetDeviceStyleQueryResult)

	result := &p_grpc.GetDeviceStyleResponse{}

	for _, ds := range styles {
		result.DeviceStyles = append(result.DeviceStyles, &p_grpc.DeviceStyle{
			Id:                 ds.ID,
			Source:             ds.Source,
			SubModel:           ds.SubModel,
			Name:               ds.Name,
			ExternalStyleId:    ds.ExternalStyleID,
			DeviceDefinitionId: ds.DeviceDefinitionID,
			HardwareTemplateId: ds.HardwareTemplateID,
		})
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceStylesByFilter(ctx context.Context, in *p_grpc.GetDeviceStyleFilterRequest) (*p_grpc.GetDeviceStyleResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByFilterQuery{
		DeviceDefinitionID: in.DeviceDefinitionId,
		Name:               in.Name,
		SubModel:           in.SubModel,
	})

	styles := qryResult.([]coremodels.GetDeviceStyleQueryResult)

	result := &p_grpc.GetDeviceStyleResponse{}

	for _, ds := range styles {
		result.DeviceStyles = append(result.DeviceStyles, &p_grpc.DeviceStyle{
			Id:                 ds.ID,
			Source:             ds.Source,
			SubModel:           ds.SubModel,
			Name:               ds.Name,
			ExternalStyleId:    ds.ExternalStyleID,
			DeviceDefinitionId: ds.DeviceDefinitionID,
			HardwareTemplateId: ds.HardwareTemplateID,
		})
	}

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceStyleByExternalID(ctx context.Context, in *p_grpc.GetDeviceStyleByIDRequest) (*p_grpc.DeviceStyle, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByExternalIDQuery{
		ExternalDeviceID: in.Id,
	})

	ds := qryResult.(coremodels.GetDeviceStyleQueryResult)
	result := &p_grpc.DeviceStyle{
		Id:                 ds.ID,
		Source:             ds.Source,
		SubModel:           ds.SubModel,
		Name:               ds.Name,
		ExternalStyleId:    ds.ExternalStyleID,
		DeviceDefinitionId: ds.DeviceDefinitionID,
		HardwareTemplateId: ds.HardwareTemplateID,
	}

	return result, nil
}

func (s *GrpcDefinitionsService) UpdateDeviceStyle(ctx context.Context, in *p_grpc.UpdateDeviceStyleRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.UpdateDeviceStyleCommand{
		ID:                 in.Id,
		Name:               in.Name,
		ExternalStyleID:    in.ExternalStyleId,
		DeviceDefinitionID: in.DeviceDefinitionId,
		Source:             in.Source,
		SubModel:           in.SubModel,
		HardwareTemplateID: in.HardwareTemplateId,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceStyleCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

//** Makes / Manufacturers

func (s *GrpcDefinitionsService) GetDeviceMakeByName(ctx context.Context, in *p_grpc.GetDeviceMakeByNameRequest) (*p_grpc.DeviceMake, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceMakeByNameQuery{
		Name: in.Name,
	})

	deviceMake := qryResult.(coremodels.DeviceMake)

	result := &p_grpc.DeviceMake{
		Id:              deviceMake.ID,
		Name:            deviceMake.Name,
		NameSlug:        deviceMake.NameSlug,
		LogoUrl:         deviceMake.LogoURL.String,
		OemPlatformName: deviceMake.OemPlatformName.String,
	}

	manufacturerID, err := s.queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, deviceMake.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to GetManufacturerIdByName for update: %s", deviceMake.Name)
	}
	result.TokenId = manufacturerID.Uint64()

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceMakeBySlug(ctx context.Context, in *p_grpc.GetDeviceMakeBySlugRequest) (*p_grpc.DeviceMake, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceMakeBySlugQuery{
		Slug: in.Slug,
	})

	deviceMake := qryResult.(coremodels.DeviceMake)

	result := &p_grpc.DeviceMake{
		Id:              deviceMake.ID,
		Name:            deviceMake.Name,
		NameSlug:        deviceMake.NameSlug,
		LogoUrl:         deviceMake.LogoURL.String,
		OemPlatformName: deviceMake.OemPlatformName.String,
	}

	manufacturerID, err := s.queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, deviceMake.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to GetManufacturerIdByName for update: %s", deviceMake.Name)
	}
	result.TokenId = manufacturerID.Uint64()

	return result, nil
}

func (s *GrpcDefinitionsService) GetDeviceMakeByTokenID(ctx context.Context, in *p_grpc.GetDeviceMakeByTokenIdRequest) (*p_grpc.DeviceMake, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceMakeByTokenIDQuery{
		TokenID: in.TokenId,
	})

	deviceMakes := qryResult.(*p_grpc.DeviceMake)

	return deviceMakes, nil
}

func (s *GrpcDefinitionsService) GetDeviceMakes(ctx context.Context, _ *emptypb.Empty) (*p_grpc.GetDeviceMakeResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceMakeQuery{})

	deviceMakes := qryResult.(*p_grpc.GetDeviceMakeResponse)

	return deviceMakes, nil
}

func (s *GrpcDefinitionsService) CreateDeviceMake(ctx context.Context, in *p_grpc.CreateDeviceMakeRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceMakeCommand{
		Name:               in.Name,
		HardwareTemplateID: in.HardwareTemplateId,
		ExternalIDs:        "{}",
		Metadata:           "{}",
	})

	result := commandResult.(commands.CreateDeviceMakeCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcDefinitionsService) UpdateDeviceMake(ctx context.Context, in *p_grpc.UpdateDeviceMakeRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.UpdateDeviceMakeCommand{
		ID:                 in.Id,
		Name:               in.Name,
		LogoURL:            null.StringFrom(in.LogoUrl),
		OemPlatformName:    null.StringFrom(in.OemPlatformName),
		ExternalIDs:        json.RawMessage(in.ExternalIds),
		Metadata:           json.RawMessage(in.Metadata),
		HardwareTemplateID: in.HardwareTemplateId,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceMakeCommandResult)

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
