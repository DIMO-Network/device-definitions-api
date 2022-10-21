package commands

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

type UpdateDeviceDefinitionCommand struct {
	DeviceDefinitionID string      `json:"deviceDefinitionId"`
	Source             null.String `json:"source"`
	ExternalID         string      `json:"external_id"`
	ImageURL           null.String `json:"image_url"`
	VehicleInfo        *UpdateDeviceVehicleInfo
	Verified           bool                       `json:"verified"`
	Model              string                     `json:"model"`
	Year               int16                      `json:"year"`
	DeviceMakeID       string                     `json:"device_make_id"`
	DeviceStyles       []UpdateDeviceStyles       `json:"deviceStyles"`
	DeviceIntegrations []UpdateDeviceIntegrations `json:"deviceIntegrations"`
	// DeviceTypeID comes from the device_types.id table, determines what kind of device this is, typically a vehicle
	DeviceTypeID string `json:"device_type_id"`
	// DeviceAttributes sets definition metadata eg. vehicle info. Allowed key/values are defined in device_types.properties
	DeviceAttributes []*coremodels.UpdateDeviceTypeAttribute `json:"deviceAttributes"`
}

type UpdateDeviceIntegrations struct {
	IntegrationID string    `json:"integration_id"`
	Capabilities  null.JSON `json:"capabilities,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	Region        string    `json:"region"`
}

type UpdateDeviceStyles struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	ExternalStyleID string    `json:"external_style_id"`
	Source          string    `json:"source"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
	SubModel        string    `json:"sub_model"`
}

type UpdateDeviceVehicleInfo struct {
	FuelType            string `json:"fuel_type,omitempty"`
	DrivenWheels        string `json:"driven_wheels,omitempty"`
	NumberOfDoors       string `json:"number_of_doors,omitempty"`
	BaseMSRP            int    `json:"base_msrp,omitempty"`
	EPAClass            string `json:"epa_class,omitempty"`
	VehicleType         string `json:"vehicle_type,omitempty"` // VehicleType PASSENGER CAR, from NHTSA
	MPGHighway          string `json:"mpg_highway,omitempty"`
	MPGCity             string `json:"mpg_city,omitempty"`
	FuelTankCapacityGal string `json:"fuel_tank_capacity_gal,omitempty"`
	MPG                 string `json:"mpg,omitempty"`
}

type UpdateDeviceDefinitionCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateDeviceDefinitionCommand) Key() string { return "UpdateDeviceDefinitionCommand" }

type UpdateDeviceDefinitionCommandHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
	DDCache    services.DeviceDefinitionCacheService
}

func NewUpdateDeviceDefinitionCommandHandler(repository repositories.DeviceDefinitionRepository, dbs func() *db.ReaderWriter, cache services.DeviceDefinitionCacheService) UpdateDeviceDefinitionCommandHandler {
	return UpdateDeviceDefinitionCommandHandler{DDCache: cache, Repository: repository, DBS: dbs}
}

func (ch UpdateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceDefinitionCommand)

	if len(command.DeviceTypeID) == 0 {
		command.DeviceTypeID = common.DefaultDeviceType
	}

	dd, err := ch.Repository.GetByID(ctx, command.DeviceDefinitionID)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	// Resolve make
	dm, err := models.DeviceMakes(models.DeviceMakeWhere.ID.EQ(command.DeviceMakeID)).One(ctx, ch.DBS().Reader)

	// Resolve attributes by device types
	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.DeviceTypeID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to get device types"),
			}
		}
	}

	if dd == nil {
		dd = &models.DeviceDefinition{
			ID:           command.DeviceDefinitionID,
			DeviceMakeID: command.DeviceMakeID,
			Model:        command.Model,
			Year:         command.Year,
			Verified:     false,
			ModelSlug:    common.SlugString(command.Model),
			DeviceTypeID: null.StringFrom(dt.ID),
		}
	}

	// check if vehicleInfo is set and load the attributes, so that we don't break existing code. In the future we may remove this if all
	// clients update to send in metadata as DeviceAttributes
	if command.VehicleInfo != nil && len(command.DeviceAttributes) == 0 {
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "fuel_type",
			Value: command.VehicleInfo.FuelType,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "driven_wheels",
			Value: command.VehicleInfo.DrivenWheels,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "number_of_doors",
			Value: command.VehicleInfo.NumberOfDoors,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "base_MSRP",
			Value: fmt.Sprintf("%d", command.VehicleInfo.BaseMSRP),
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "EPA_class",
			Value: command.VehicleInfo.EPAClass,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "vehicle_type",
			Value: command.VehicleInfo.VehicleType,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "MPG_city",
			Value: command.VehicleInfo.MPGCity,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "MPG_highway",
			Value: command.VehicleInfo.MPGHighway,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "MPG",
			Value: command.VehicleInfo.MPG,
		})
		command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
			Name:  "fuel_tank_capacity_gal",
			Value: command.VehicleInfo.FuelTankCapacityGal,
		})
	}

	// attribute info
	deviceTypeInfo, err := common.BuildDeviceTypeAttributes(command.DeviceAttributes, dt)
	if err != nil {
		return nil, err
	}

	if len(command.Model) > 0 {
		dd.Model = command.Model
	}

	if command.Year > 0 {
		dd.Year = command.Year
	}

	if len(command.DeviceMakeID) > 0 {
		dd.DeviceMakeID = command.DeviceMakeID
	}

	if len(command.ExternalID) > 0 {
		dd.ExternalID = null.StringFrom(command.ExternalID)
	}

	dd.Source = command.Source
	dd.ImageURL = command.ImageURL
	dd.Verified = command.Verified

	var deviceStyles []*models.DeviceStyle
	var deviceIntegrations []*models.DeviceIntegration

	if len(command.DeviceStyles) > 0 {
		for _, ds := range command.DeviceStyles {
			deviceStyles = append(deviceStyles, &models.DeviceStyle{
				ID:                 ds.ID,
				DeviceDefinitionID: command.DeviceDefinitionID,
				Name:               ds.Name,
				ExternalStyleID:    ds.ExternalStyleID,
				Source:             ds.Source,
				CreatedAt:          ds.CreatedAt,
				UpdatedAt:          ds.UpdatedAt,
				SubModel:           ds.SubModel,
			})
		}
	}

	if len(command.DeviceIntegrations) > 0 {
		for _, di := range command.DeviceIntegrations {
			deviceIntegrations = append(deviceIntegrations, &models.DeviceIntegration{
				DeviceDefinitionID: command.DeviceDefinitionID,
				IntegrationID:      di.IntegrationID,
				CreatedAt:          di.CreatedAt,
				UpdatedAt:          di.UpdatedAt,
				Region:             di.Region,
			})
		}
	}

	dd, err = ch.Repository.CreateOrUpdate(ctx, dd, deviceStyles, deviceIntegrations, deviceTypeInfo)

	if err != nil {
		return nil, err
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, dd.ID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dm.Name, dd.Model, int(dd.Year))

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
