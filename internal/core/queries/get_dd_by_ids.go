package queries

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetDeviceDefinitionByIdsQuery struct {
	DeviceDefinitionID []string `json:"deviceDefinitionId" validate:"required"`
}

func (*GetDeviceDefinitionByIdsQuery) Key() string { return "GetDeviceDefinitionByIdsQuery" }

type GetDeviceDefinitionByIdsQueryHandler struct {
	DDCache services.DeviceDefinitionCacheService
	log     *zerolog.Logger
}

func NewGetDeviceDefinitionByIdsQueryHandler(cache services.DeviceDefinitionCacheService, log *zerolog.Logger) GetDeviceDefinitionByIdsQueryHandler {
	return GetDeviceDefinitionByIdsQueryHandler{
		DDCache: cache,
		log:     log,
	}
}

func (ch GetDeviceDefinitionByIdsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByIdsQuery)

	if len(qry.DeviceDefinitionID) == 0 {
		return nil, &exceptions.ValidationError{
			Err: errors.New("Device Definition Ids is required"),
		}
	}

	response := &grpc.GetDeviceDefinitionResponse{}

	for _, v := range qry.DeviceDefinitionID {
		dd, _ := ch.DDCache.GetDeviceDefinitionByID(ctx, v)

		if dd == nil {
			if len(qry.DeviceDefinitionID) > 1 {
				ch.log.Warn().Str("deviceDefinitionId", v).Msg("Not found - Device Definition")
				continue
			}

			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find device definition id: %s", v),
			}
		}

		rp := &grpc.GetDeviceDefinitionItemResponse{
			DeviceDefinitionId: dd.DeviceDefinitionID,
			Name:               dd.Name,
			ImageUrl:           dd.ImageURL,
			Source:             dd.Source,
			Make: &grpc.DeviceMake{
				Id:              dd.DeviceMake.ID,
				Name:            dd.DeviceMake.Name,
				LogoUrl:         dd.DeviceMake.LogoURL.String,
				OemPlatformName: dd.DeviceMake.OemPlatformName.String,
				NameSlug:        dd.DeviceMake.NameSlug,
			},
			Type: &grpc.DeviceType{
				Type:      dd.Type.Type,
				Make:      dd.DeviceMake.Name,
				Model:     dd.Type.Model,
				Year:      int32(dd.Type.Year),
				MakeSlug:  dd.Type.MakeSlug,
				ModelSlug: dd.Type.ModelSlug,
			},
			Verified: dd.Verified,
		}

		if dd.DeviceMake.TokenID != nil {
			rp.Make.TokenId = dd.DeviceMake.TokenID.Uint64()
		}

		// vehicle info
		numberOfDoors, _ := strconv.ParseInt(dd.VehicleInfo.NumberOfDoors, 6, 12)
		mpgHighway, _ := strconv.ParseFloat(dd.VehicleInfo.MPGHighway, 32)
		mpgCity, _ := strconv.ParseFloat(dd.VehicleInfo.MPGCity, 32)
		fuelTankCapacityGal, _ := strconv.ParseFloat(dd.VehicleInfo.FuelTankCapacityGal, 32)
		mpg, _ := strconv.ParseFloat(dd.VehicleInfo.MPG, 32)

		rp.VehicleData = &grpc.VehicleInfo{
			FuelType:            dd.VehicleInfo.FuelType,
			DrivenWheels:        dd.VehicleInfo.DrivenWheels,
			NumberOfDoors:       int32(numberOfDoors),
			Base_MSRP:           int32(dd.VehicleInfo.BaseMSRP),
			EPAClass:            dd.VehicleInfo.EPAClass,
			VehicleType:         dd.VehicleInfo.VehicleType,
			MPGHighway:          float32(mpgHighway),
			MPGCity:             float32(mpgCity),
			FuelTankCapacityGal: float32(fuelTankCapacityGal),
			MPG:                 float32(mpg),
		}

		// sub_models
		rp.Type.SubModels = dd.Type.SubModels

		// build object for integrations that have all the info
		rp.DeviceIntegrations = []*grpc.DeviceIntegration{}
		for _, di := range dd.DeviceIntegrations {
			rp.DeviceIntegrations = append(rp.DeviceIntegrations, &grpc.DeviceIntegration{
				DeviceDefinitionId: dd.DeviceDefinitionID,
				Integration: &grpc.Integration{
					Id:     di.ID,
					Type:   di.Type,
					Style:  di.Style,
					Vendor: di.Vendor,
				},
				Region:       di.Region,
				Capabilities: string(di.Capabilities),
			})
		}

		rp.DeviceStyles = []*grpc.DeviceStyles{}
		for _, ds := range dd.DeviceStyles {
			rp.DeviceStyles = append(rp.DeviceStyles, &grpc.DeviceStyles{
				DeviceDefinitionId: dd.DeviceDefinitionID,
				ExternalStyleId:    ds.ExternalStyleID,
				Id:                 ds.ID,
				Name:               ds.Name,
				Source:             ds.Source,
				SubModel:           ds.SubModel,
			})
		}

		response.DeviceDefinitions = append(response.DeviceDefinitions, rp)
	}

	return response, nil
}
