package api

import (
	"context"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
)

type GrpcService struct {
	p_grpc.UnimplementedDeviceDefinitionServiceServer
	Mediator mediator.Mediator
}

func NewGrpcService(mediator mediator.Mediator) p_grpc.DeviceDefinitionServiceServer {
	return &GrpcService{Mediator: mediator}
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

	dd := qryResult.(queries.GetDeviceDefinitionByMakeModelYearQueryResult)

	numberOfDoors, _ := strconv.ParseInt(dd.VehicleInfo.NumberOfDoors, 6, 12)
	mpgHighway, _ := strconv.ParseFloat(dd.VehicleInfo.MPGHighway, 32)
	mpgCity, _ := strconv.ParseFloat(dd.VehicleInfo.MPGCity, 32)
	fuelTankCapacityGal, _ := strconv.ParseFloat(dd.VehicleInfo.FuelTankCapacityGal, 32)

	result := &p_grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: dd.DeviceDefinitionID,
		Name:               dd.Name,
		ImageUrl:           dd.ImageURL,
		Type: &p_grpc.GetDeviceDefinitionItemResponse_Type{
			Type:  dd.Type.Type,
			Make:  dd.Type.Make,
			Model: dd.Type.Model,
			Year:  int32(dd.Type.Year),
		},
		VehicleData: &p_grpc.GetDeviceDefinitionItemResponse_VehicleInfo{
			FuelType:            dd.VehicleInfo.FuelType,
			DrivenWheels:        dd.VehicleInfo.DrivenWheels,
			NumberOfDoors:       int32(numberOfDoors),
			Base_MSRP:           int32(dd.VehicleInfo.BaseMSRP),
			EPAClass:            dd.VehicleInfo.EPAClass,
			VehicleType:         dd.VehicleInfo.VehicleType,
			MPGHighway:          float32(mpgHighway),
			MPGCity:             float32(mpgCity),
			FuelTankCapacityGal: float32(fuelTankCapacityGal),
		},
		Verified: dd.Verified,
	}

	for _, integration := range dd.CompatibleIntegrations {
		result.DeviceIntegrations = append(result.DeviceIntegrations, &p_grpc.GetDeviceDefinitionItemResponse_DeviceIntegrations{
			Id:      integration.ID,
			Type:    integration.Type,
			Style:   integration.Style,
			Vendor:  integration.Vendor,
			Region:  integration.Region,
			Country: integration.Country,
		})
	}

	return result, nil
}

func (s *GrpcService) GetIntegrations(ctx context.Context, in *p_grpc.EmptyRequest) (*p_grpc.GetIntegrationResponse, error) {

	qryResult, _ := s.Mediator.Send(ctx, &queries.GetAllIntegrationQuery{})

	integrations := qryResult.([]queries.GetAllIntegrationQueryResult)
	result := &p_grpc.GetIntegrationResponse{}

	for _, item := range integrations {
		result.Integrations = append(result.Integrations, &p_grpc.GetIntegrationItemResponse{
			Id:     item.ID,
			Type:   item.Type,
			Style:  item.Style,
			Vendor: item.Vendor,
		})
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
		result.Integrations = append(result.Integrations, &p_grpc.GetDeviceDefinitionIntegrationItemResponse{
			Id:      queryResult.ID,
			Type:    queryResult.Type,
			Style:   queryResult.Style,
			Vendor:  queryResult.Vendor,
			Region:  queryResult.Region,
			Country: queryResult.Country,
		})
	}

	return result, nil
}

func (s *GrpcService) CreateDeviceDefinition(ctx context.Context, in *p_grpc.CreateDeviceDefinitionRequest) (*p_grpc.CreateDeviceDefinitionResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceDefinitionCommand{
		Source: in.Source,
		Make:   in.Make,
		Model:  in.Model,
		Year:   int(in.Year),
	})

	result := commandResult.(commands.CreateDeviceDefinitionCommandResult)

	return &p_grpc.CreateDeviceDefinitionResponse{Id: result.ID}, nil
}

func (s *GrpcService) CreateDeviceIntegration(ctx context.Context, in *p_grpc.CreateDeviceIntegrationRequest) (*p_grpc.CreateDeviceIntegrationResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceIntegrationCommand{
		DeviceDefinitionID: in.DeviceDefinitionId,
		IntegrationID:      in.IntegrationId,
		Region:             in.Region,
	})

	result := commandResult.(commands.CreateDeviceIntegrationCommandResult)

	return &p_grpc.CreateDeviceIntegrationResponse{Id: result.ID}, nil
}

func (s *GrpcService) CreateDeviceIntegration(ctx context.Context, in *p_grpc.CreateDeviceIntegrationRequest) (*p_grpc.CreateDeviceIntegrationResponse, error) {

	commandResult, _ := s.Mediator.Send(ctx, &commands.CreateDeviceIntegrationCommand{
		DeviceDefinitionId: in.DeviceDefinitionId,
		IntegrationId:      in.IntegrationId,
		Region:             in.Region,
	})

	result := commandResult.(commands.CreateDeviceIntegrationCommandResult)

	return &p_grpc.CreateDeviceIntegrationResponse{Id: result.Id}, nil
}
