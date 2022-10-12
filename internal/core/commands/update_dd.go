package commands

import (
	"context"
	"database/sql"

	"fmt"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UpdateDeviceDefinitionCommand struct {
	DeviceDefinitionID string      `json:"deviceDefinitionId"`
	Source             null.String `json:"source"`
	ExternalID         string      `json:"external_id"`
	ImageURL           null.String `json:"image_url"`
	VehicleInfo        UpdateDeviceVehicleInfo
	Verified           bool                       `json:"verified"`
	Model              string                     `json:"model"`
	Year               int16                      `json:"year"`
	DeviceMakeID       string                     `json:"device_make_id"`
	DeviceStyles       []UpdateDeviceStyles       `json:"deviceStyles"`
	DeviceIntegrations []UpdateDeviceIntegrations `json:"deviceIntegrations"`
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
	DBS     func() *db.ReaderWriter
	DDCache services.DeviceDefinitionCacheService
}

func NewUpdateDeviceDefinitionCommandHandler(dbs func() *db.ReaderWriter, cache services.DeviceDefinitionCacheService) UpdateDeviceDefinitionCommandHandler {
	return UpdateDeviceDefinitionCommandHandler{DBS: dbs, DDCache: cache}
}

func (ch UpdateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceDefinitionCommand)

	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", command.DeviceDefinitionID),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration))).
		One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if err != nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", command.DeviceDefinitionID),
		}
	}

	// Update Vehicle Info
	deviceVehicleInfoMetaData := new(UpdateDeviceVehicleInfo)
	if err := dd.Metadata.Unmarshal(deviceVehicleInfoMetaData); err == nil {
		deviceVehicleInfoMetaData.FuelType = command.VehicleInfo.FuelType
		deviceVehicleInfoMetaData.DrivenWheels = command.VehicleInfo.DrivenWheels
		deviceVehicleInfoMetaData.NumberOfDoors = command.VehicleInfo.NumberOfDoors
		deviceVehicleInfoMetaData.BaseMSRP = int(command.VehicleInfo.BaseMSRP)
		deviceVehicleInfoMetaData.EPAClass = command.VehicleInfo.EPAClass
		deviceVehicleInfoMetaData.VehicleType = command.VehicleInfo.VehicleType
		deviceVehicleInfoMetaData.MPGCity = command.VehicleInfo.MPGCity
		deviceVehicleInfoMetaData.MPGHighway = command.VehicleInfo.MPGHighway
		deviceVehicleInfoMetaData.MPG = command.VehicleInfo.MPG
		deviceVehicleInfoMetaData.FuelTankCapacityGal = command.VehicleInfo.FuelTankCapacityGal
	}

	err = dd.Metadata.Marshal(deviceVehicleInfoMetaData)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
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

	_, err = dd.Update(ctx, ch.DBS().Writer.DB, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	if len(command.DeviceStyles) > 0 {
		// Remove Device Styles
		_, err = models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(command.DeviceDefinitionID)).
			DeleteAll(ctx, ch.DBS().Writer.DB)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		// Update Device Styles
		for _, ds := range command.DeviceStyles {
			subModels := &models.DeviceStyle{
				ID:                 ds.ID,
				DeviceDefinitionID: command.DeviceDefinitionID,
				Name:               ds.Name,
				ExternalStyleID:    ds.ExternalStyleID,
				Source:             ds.Source,
				CreatedAt:          ds.CreatedAt,
				UpdatedAt:          ds.UpdatedAt,
				SubModel:           ds.SubModel,
			}
			err = subModels.Insert(ctx, ch.DBS().Writer.DB, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: err,
				}
			}
		}
	}

	if len(command.DeviceIntegrations) > 0 {
		// Remove Device Integrations
		_, err = models.DeviceIntegrations(models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(command.DeviceDefinitionID)).
			DeleteAll(ctx, ch.DBS().Writer.DB)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		for _, di := range command.DeviceIntegrations {
			deviceIntegration := &models.DeviceIntegration{
				DeviceDefinitionID: command.DeviceDefinitionID,
				IntegrationID:      di.IntegrationID,
				//Capabilities:       di.Capabilities,
				CreatedAt: di.CreatedAt,
				UpdatedAt: di.UpdatedAt,
				Region:    di.Region,
			}
			err = deviceIntegration.Insert(ctx, ch.DBS().Writer.DB, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: err,
				}
			}
		}
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, command.DeviceDefinitionID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year))

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
