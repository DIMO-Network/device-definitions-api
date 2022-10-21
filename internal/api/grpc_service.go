package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	elasticModels "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elasticsearch/models"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type GrpcService struct {
	p_grpc.UnimplementedDeviceDefinitionServiceServer
	Mediator mediator.Mediator
	logger   *zerolog.Logger
}

func NewGrpcService(mediator mediator.Mediator, logger *zerolog.Logger) p_grpc.DeviceDefinitionServiceServer {
	return &GrpcService{Mediator: mediator, logger: logger}
}

func (s *GrpcService) GetDeviceDefinitionByID(ctx context.Context, in *p_grpc.GetDeviceDefinitionRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByIdsQuery{
		DeviceDefinitionID: in.Ids,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcService) GetDeviceDefinitionByMMY(ctx context.Context, in *p_grpc.GetDeviceDefinitionByMMYRequest) (*p_grpc.GetDeviceDefinitionItemResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionByMakeModelYearQuery{
		Make:  in.Make,
		Model: in.Model,
		Year:  int(in.Year),
	})

	dd := qryResult.(*models.GetDeviceDefinitionQueryResult)

	//nolint
	numberOfDoors, _ := strconv.ParseInt(dd.VehicleInfo.NumberOfDoors, 6, 12)
	//nolint
	mpgHighway, _ := strconv.ParseFloat(dd.VehicleInfo.MPGHighway, 32)
	//nolint
	mpgCity, _ := strconv.ParseFloat(dd.VehicleInfo.MPGCity, 32)
	//nolint
	mpg, _ := strconv.ParseFloat(dd.VehicleInfo.MPG, 32)
	//nolint
	fuelTankCapacityGal, _ := strconv.ParseFloat(dd.VehicleInfo.FuelTankCapacityGal, 32)

	result := &p_grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: dd.DeviceDefinitionID,
		Name:               dd.Name,
		ImageUrl:           dd.ImageURL,
		Source:             dd.Source,
		Type: &p_grpc.DeviceType{
			Type:      dd.Type.Type,
			Make:      dd.Type.Make,
			Model:     dd.Type.Model,
			Year:      int32(dd.Type.Year),
			MakeSlug:  dd.Type.MakeSlug,
			ModelSlug: dd.Type.ModelSlug,
			SubModels: dd.Type.SubModels,
		},
		Make: &p_grpc.DeviceMake{
			Id:              dd.DeviceMake.ID,
			Name:            dd.DeviceMake.Name,
			LogoUrl:         dd.DeviceMake.LogoURL.String,
			OemPlatformName: dd.DeviceMake.OemPlatformName.String,
			NameSlug:        dd.DeviceMake.NameSlug,
		},
		//nolint
		VehicleData: &p_grpc.VehicleInfo{
			FuelType: dd.VehicleInfo.FuelType,
			DrivenWheels:  dd.VehicleInfo.DrivenWheels,
			NumberOfDoors: int32(numberOfDoors),
			Base_MSRP: int32(dd.VehicleInfo.BaseMSRP),
			EPAClass: dd.VehicleInfo.EPAClass,
			VehicleType:         dd.VehicleInfo.VehicleType,
			MPGHighway:          float32(mpgHighway),
			MPGCity:             float32(mpgCity),
			FuelTankCapacityGal: float32(fuelTankCapacityGal),
			MPG:                 float32(mpg),
		},
		Verified: dd.Verified,
	}
	if dd.DeviceMake.TokenID != nil {
		result.Make.TokenId = dd.DeviceMake.TokenID.Uint64()
	}

	for _, integration := range dd.DeviceIntegrations {
		result.DeviceIntegrations = append(result.DeviceIntegrations, &p_grpc.DeviceIntegration{
			Integration: &p_grpc.Integration{
				Id:     integration.ID,
				Type:   integration.Type,
				Style:  integration.Style,
				Vendor: integration.Vendor,
			},
			Region:             integration.Region,
			DeviceDefinitionId: dd.DeviceDefinitionID,
		})
	}

	return result, nil
}

func (s *GrpcService) GetFilteredDeviceDefinition(ctx context.Context, in *p_grpc.FilterDeviceDefinitionRequest) (*p_grpc.GetFilteredDeviceDefinitionsResponse, error) {
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
		result.Items = append(result.Items, &p_grpc.FilterDeviceDefinitionsReponse{
			Id:           deviceDefinition.ID,
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
		})
	}

	return result, nil
}

func (s *GrpcService) GetDeviceDefinitionBySource(ctx context.Context, in *p_grpc.GetDeviceDefinitionBySourceRequest) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionBySourceQuery{
		Source: in.Source,
	})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcService) GetDeviceDefinitionWithoutImages(ctx context.Context, in *emptypb.Empty) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceDefinitionWithoutImageQuery{})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcService) GetIntegrations(ctx context.Context, in *emptypb.Empty) (*p_grpc.GetIntegrationResponse, error) {

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

func (s *GrpcService) GetIntegrationByID(ctx context.Context, in *p_grpc.GetIntegrationRequest) (*p_grpc.Integration, error) {

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

func (s *GrpcService) GetDeviceDefinitionIntegration(ctx context.Context, in *p_grpc.GetDeviceDefinitionIntegrationRequest) (*p_grpc.GetDeviceDefinitionIntegrationResponse, error) {

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

func (s *GrpcService) CreateDeviceDefinition(ctx context.Context, in *p_grpc.CreateDeviceDefinitionRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceDefinitionCommand{
		Source: in.Source,
		Make:   in.Make,
		Model:  in.Model,
		Year:   int(in.Year),
	})

	result := commandResult.(commands.CreateDeviceDefinitionCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) CreateDeviceIntegration(ctx context.Context, in *p_grpc.CreateDeviceIntegrationRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceIntegrationCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		IntegrationID:      in.IntegrationId,
		Region:             in.Region,
	})

	result := commandResult.(commands.CreateDeviceIntegrationCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) CreateDeviceStyle(ctx context.Context, in *p_grpc.CreateDeviceStyleRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceStyleCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		Name:               in.Name,
		ExternalStyleID:    in.ExternalStyleId,
		Source:             in.Source,
		SubModel:           in.SubModel,
	})

	result := commandResult.(commands.CreateDeviceStyleCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) GetDeviceMakeByName(ctx context.Context, in *p_grpc.GetDeviceMakeByNameRequest) (*p_grpc.DeviceMake, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceMakeByNameQuery{
		Name: in.Name,
	})

	deviceMake := qryResult.(models.DeviceMake)

	result := &p_grpc.DeviceMake{
		Id:              deviceMake.ID,
		Name:            deviceMake.Name,
		NameSlug:        deviceMake.NameSlug,
		LogoUrl:         deviceMake.LogoURL.String,
		OemPlatformName: deviceMake.OemPlatformName.String,
		ExternalIds:     string(deviceMake.ExternalIds),
	}

	if deviceMake.TokenID != nil {
		result.TokenId = deviceMake.TokenID.Uint64()
	}

	return result, nil
}

func (s *GrpcService) GetDeviceMakes(ctx context.Context, in *emptypb.Empty) (*p_grpc.GetDeviceMakeResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceMakeQuery{})

	deviceMakes := qryResult.([]models.DeviceMake)

	result := &p_grpc.GetDeviceMakeResponse{}

	for _, deviceMake := range deviceMakes {

		make := &p_grpc.DeviceMake{
			Id:              deviceMake.ID,
			Name:            deviceMake.Name,
			NameSlug:        deviceMake.NameSlug,
			LogoUrl:         deviceMake.LogoURL.String,
			OemPlatformName: deviceMake.OemPlatformName.String,
			ExternalIds:     string(deviceMake.ExternalIds),
		}

		if deviceMake.TokenID != nil {
			make.TokenId = deviceMake.TokenID.Uint64()
		}

		result.DeviceMakes = append(result.DeviceMakes, make)
	}

	return result, nil
}

func (s *GrpcService) CreateDeviceMake(ctx context.Context, in *p_grpc.CreateDeviceMakeRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceMakeCommand{
		Name: in.Name,
	})

	result := commandResult.(commands.CreateDeviceMakeCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) CreateIntegration(ctx context.Context, in *p_grpc.CreateIntegrationRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateIntegrationCommand{
		Vendor: in.Vendor,
		Style:  in.Style,
		Type:   in.Type,
	})

	result := commandResult.(commands.CreateIntegrationCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) UpdateDeviceDefinition(ctx context.Context, in *p_grpc.UpdateDeviceDefinitionRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.UpdateDeviceDefinitionCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		Source:             null.StringFrom(in.Source),
		ExternalID:         in.ExternalId,
		ImageURL:           null.StringFrom(in.ImageUrl),
		Year:               int16(in.Year),
		Model:              in.Model,
		Verified:           in.Verified,
		DeviceMakeID:       in.DeviceMakeId,
	}

	//nolint
	if in.VehicleData != nil {
		command.VehicleInfo = &commands.UpdateDeviceVehicleInfo{
			FuelType:            in.VehicleData.FuelType,
			DrivenWheels:        in.VehicleData.DrivenWheels,
			NumberOfDoors:       strconv.Itoa(int(in.VehicleData.NumberOfDoors)),
			BaseMSRP:            int(in.VehicleData.Base_MSRP),
			EPAClass:            in.VehicleData.EPAClass,
			VehicleType:         in.VehicleData.VehicleType,
			MPGHighway:          fmt.Sprintf("%f", in.VehicleData.MPGHighway),
			FuelTankCapacityGal: fmt.Sprintf("%f", in.VehicleData.FuelTankCapacityGal),
			MPGCity:             fmt.Sprintf("%f", in.VehicleData.MPGCity),
			MPG:                 fmt.Sprintf("%f", in.VehicleData.MPG),
		}
	}

	if len(in.DeviceStyles) > 0 {
		for _, style := range in.DeviceStyles {
			command.DeviceStyles = append(command.DeviceStyles, commands.UpdateDeviceStyles{
				ID:              style.Id,
				ExternalStyleID: style.ExternalStyleId,
				Name:            style.Name,
				Source:          style.Source,
				SubModel:        style.SubModel,
				CreatedAt:       style.CreatedAt.AsTime(),
				UpdatedAt:       style.UpdatedAt.AsTime(),
			})
		}
	}

	if len(in.DeviceIntegrations) > 0 {
		for _, integration := range in.DeviceIntegrations {
			command.DeviceIntegrations = append(command.DeviceIntegrations, commands.UpdateDeviceIntegrations{
				IntegrationID: integration.IntegrationId,
				Region:        integration.Region,
				CreatedAt:     integration.CreatedAt.AsTime(),
				UpdatedAt:     integration.UpdatedAt.AsTime(),
			})
		}
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceDefinitionCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) SetDeviceDefinitionImage(ctx context.Context, in *p_grpc.UpdateDeviceDefinitionImageRequest) (*p_grpc.BaseResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.UpdateDeviceDefinitionImageCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		ImageURL:           in.ImageUrl,
	})

	result := commandResult.(commands.CreateDeviceDefinitionCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) GetDeviceDefinitionAll(ctx context.Context, in *emptypb.Empty) (*p_grpc.GetDeviceDefinitionAllResponse, error) {

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

func (s *GrpcService) GetDeviceCompatibilities(ctx context.Context, in *p_grpc.GetDeviceCompatibilityListRequest) (*p_grpc.GetDeviceCompatibilityListResponse, error) {
	logger := s.logger.With().Str("rpc", "GetDeviceCompatibilities").Logger()
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceCompatibilityQuery{
		MakeID:        in.MakeId,
		IntegrationID: in.IntegrationId,
		Region:        in.Region,
	})

	deviceCompatibilities := qryResult.(queries.GetDeviceCompatibilityQueryResult)

	logger.Debug().Interface("queryResult", deviceCompatibilities).Msg("Completed compatibility query.")

	result := &p_grpc.GetDeviceCompatibilityListResponse{}

	integFeats := deviceCompatibilities.IntegrationFeatures
	totalWeightsCount := 0.0
	for _, v := range deviceCompatibilities.IntegrationFeatures {
		totalWeightsCount += v.FeatureWeight
	}
	dcMap := make(map[string][]*p_grpc.DeviceCompatibilities)

	// Group by model name.
	for _, v := range deviceCompatibilities.DeviceDefinitions {
		logger := logger.With().Str("modelName", v.Model).Str("deviceDefinitionId", v.ID).Logger()
		if len(v.R.DeviceIntegrations) == 0 {
			// This should never happen, because of the inner join.
			logger.Error().Msg("No integrations for this definition.")
			continue
		}

		di := v.R.DeviceIntegrations[0]

		logger.Debug().Interface("deviceIntegration", di).Msg("Loaded device integration.")

		if di.Features.IsZero() {
			// This should never happen, because we filtered for "features IS NOT NULL".
			logger.Error().Msg("Feature column was null.")
			continue
		}
		res := &p_grpc.DeviceCompatibilities{Year: int32(v.Year)}

		feats := []*p_grpc.Feature{}
		var features []elasticModels.DeviceIntegrationFeatures

		err := di.Features.Unmarshal(&features)
		if err != nil {
			logger.Err(err).Msg("Error unmarshaling features JSON blob.")
			continue
		}

		ifeat := map[string]queries.FeatureDetails{}

		for _, f := range features {
			ft := &p_grpc.Feature{
				Key:          integFeats[f.FeatureKey].DisplayName,
				CssIcon:      integFeats[f.FeatureKey].CSSIcon,
				SupportLevel: int32(f.SupportLevel),
			}

			fts := queries.FeatureDetails{
				FeatureWeight: integFeats[f.FeatureKey].FeatureWeight,
				SupportLevel:  int32(f.SupportLevel),
			}

			ifeat[f.FeatureKey] = fts

			feats = append(feats, ft)
		}

		level := queries.GetDeviceCompatibilityLevel(ifeat, totalWeightsCount)
		res.Features = feats
		res.Level = level
		dcMap[v.Model] = append(dcMap[v.Model], res)
	}

	for k, v := range dcMap {
		dcr := &p_grpc.DeviceCompatibilityList{Name: k, Years: v}
		result.Models = append(result.Models, dcr)
	}

	return result, nil
}

func (s *GrpcService) GetDeviceDefinitions(ctx context.Context, in *emptypb.Empty) (*p_grpc.GetDeviceDefinitionResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllDeviceDefinitionQuery{})

	result := qryResult.(*p_grpc.GetDeviceDefinitionResponse)

	return result, nil
}

func (s *GrpcService) GetDeviceStyleByID(ctx context.Context, in *p_grpc.GetDeviceStyleByIDRequest) (*p_grpc.DeviceStyle, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByIDQuery{
		DeviceStyleID: in.Id,
	})

	ds := qryResult.(models.GetDeviceStyleQueryResult)
	result := &p_grpc.DeviceStyle{
		Id:                 ds.ID,
		Source:             ds.Source,
		SubModel:           ds.SubModel,
		Name:               ds.Name,
		ExternalStyleId:    ds.ExternalStyleID,
		DeviceDefinitionId: ds.DeviceDefinitionID,
	}

	return result, nil
}

func (s *GrpcService) GetDeviceStylesByDeviceDefinitionID(ctx context.Context, in *p_grpc.GetDeviceStyleByDeviceDefinitionIDRequest) (*p_grpc.GetDeviceStyleResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByDeviceDefinitionIDQuery{
		DeviceDefinitionID: in.Id,
	})

	styles := qryResult.([]models.GetDeviceStyleQueryResult)

	result := &p_grpc.GetDeviceStyleResponse{}

	for _, ds := range styles {
		result.DeviceStyles = append(result.DeviceStyles, &p_grpc.DeviceStyle{
			Id:                 ds.ID,
			Source:             ds.Source,
			SubModel:           ds.SubModel,
			Name:               ds.Name,
			ExternalStyleId:    ds.ExternalStyleID,
			DeviceDefinitionId: ds.DeviceDefinitionID,
		})
	}

	return result, nil
}

func (s *GrpcService) GetDeviceStylesByFilter(ctx context.Context, in *p_grpc.GetDeviceStyleFilterRequest) (*p_grpc.GetDeviceStyleResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByFilterQuery{
		DeviceDefinitionID: in.DeviceDefinitionId,
		Name:               in.Name,
		SubModel:           in.SubModel,
	})

	styles := qryResult.([]models.GetDeviceStyleQueryResult)

	result := &p_grpc.GetDeviceStyleResponse{}

	for _, ds := range styles {
		result.DeviceStyles = append(result.DeviceStyles, &p_grpc.DeviceStyle{
			Id:                 ds.ID,
			Source:             ds.Source,
			SubModel:           ds.SubModel,
			Name:               ds.Name,
			ExternalStyleId:    ds.ExternalStyleID,
			DeviceDefinitionId: ds.DeviceDefinitionID,
		})
	}

	return result, nil
}

func (s *GrpcService) GetDeviceStyleByExternalID(ctx context.Context, in *p_grpc.GetDeviceStyleByIDRequest) (*p_grpc.DeviceStyle, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceStyleByExternalIDQuery{
		ExternalDeviceID: in.Id,
	})

	ds := qryResult.(models.GetDeviceStyleQueryResult)
	result := &p_grpc.DeviceStyle{
		Id:                 ds.ID,
		Source:             ds.Source,
		SubModel:           ds.SubModel,
		Name:               ds.Name,
		ExternalStyleId:    ds.ExternalStyleID,
		DeviceDefinitionId: ds.DeviceDefinitionID,
	}

	return result, nil
}

func (s *GrpcService) UpdateDeviceMake(ctx context.Context, in *p_grpc.UpdateDeviceMakeRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.UpdateDeviceMakeCommand{
		ID:              in.Id,
		Name:            in.Name,
		LogoURL:         null.StringFrom(in.LogoUrl),
		OemPlatformName: null.StringFrom(in.OemPlatformName),
		ExternalIds:     json.RawMessage(in.ExternalIds),
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceMakeCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) UpdateDeviceStyle(ctx context.Context, in *p_grpc.UpdateDeviceStyleRequest) (*p_grpc.BaseResponse, error) {

	command := &commands.UpdateDeviceStyleCommand{
		ID:                 in.Id,
		Name:               in.Name,
		ExternalStyleID:    in.ExternalStyleId,
		DeviceDefinitionID: in.DeviceDefinitionId,
		Source:             in.Source,
		SubModel:           in.SubModel,
	}

	commandResult, _ := s.Mediator.Send(ctx, command)

	result := commandResult.(commands.UpdateDeviceStyleCommandResult)

	return &p_grpc.BaseResponse{Id: result.ID}, nil
}

func (s *GrpcService) GetDeviceTypesByID(ctx context.Context, in *p_grpc.GetDeviceTypeByIDRequest) (*p_grpc.GetDeviceTypeResponse, error) {
	qryResult, _ := s.Mediator.Send(ctx, &queries.GetDeviceTypeByIDQuery{
		DeviceTypeID: in.Id,
	})

	dt := qryResult.(models.GetDeviceTypeQueryResult)
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
			Options:      prop.Option,
		})
	}

	return result, nil
}
