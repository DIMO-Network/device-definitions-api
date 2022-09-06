package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UpdateDeviceDefinitionCommand struct {
	DeviceDefinitionID string `json:"deviceDefinitionId"`
	VehicleInfo        UpdateDeviceVehicleInfo
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
			return nil, &common.InternalError{
				Err: err,
			}
		}
	}

	if err != nil {
		return nil, &common.NotFoundError{
			Err: fmt.Errorf("could not find device definition id: %s", command.DeviceDefinitionID),
		}
	}

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
		return nil, &common.InternalError{
			Err: err,
		}
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, command.DeviceDefinitionID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year))

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
