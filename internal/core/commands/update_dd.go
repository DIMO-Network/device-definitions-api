package commands

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UpdateDeviceDefinitionCommand struct {
	DeviceDefinitionID  string  `json:"deviceDefinitionId"`
	FuelType            string  `json:"fuel_type"`
	DrivenWheels        string  `json:"driven_wheels"`
	NumberOfDoors       int32   `json:"number_of_doors"`
	BaseMSRP            int32   `json:"base_msrp"`
	EPAClass            string  `json:"epa_class"`
	VehicleType         string  `json:"vehicle_type"` // VehicleType PASSENGER CAR, from NHTSA
	MPGHighway          float32 `json:"mpg_highway"`
	MPGCity             float32 `json:"mpg_city"`
	FuelTankCapacityGal float32 `json:"fuel_tank_capacity_gal"`
	MPG                 float32 `json:"mpg,omitempty"`
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
	DBS func() *db.ReaderWriter
}

func NewUpdateDeviceDefinitionCommandHandler(dbs func() *db.ReaderWriter) UpdateDeviceDefinitionCommandHandler {
	return UpdateDeviceDefinitionCommandHandler{DBS: dbs}
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
		deviceVehicleInfoMetaData.FuelType = command.FuelType
		deviceVehicleInfoMetaData.DrivenWheels = command.DrivenWheels
		deviceVehicleInfoMetaData.NumberOfDoors = strconv.Itoa(int(command.NumberOfDoors))
		deviceVehicleInfoMetaData.BaseMSRP = int(command.BaseMSRP)
		deviceVehicleInfoMetaData.EPAClass = command.EPAClass
		deviceVehicleInfoMetaData.VehicleType = command.VehicleType
		deviceVehicleInfoMetaData.MPGCity = fmt.Sprintf("%f", command.MPGCity)
		deviceVehicleInfoMetaData.MPGHighway = fmt.Sprintf("%f", command.MPGHighway)
		deviceVehicleInfoMetaData.MPG = fmt.Sprintf("%f", command.MPG)
		deviceVehicleInfoMetaData.FuelTankCapacityGal = fmt.Sprintf("%f", command.FuelTankCapacityGal)
	}

	err = dd.Metadata.Marshal(deviceVehicleInfoMetaData)
	if err != nil {
		return nil, &common.InternalError{
			Err: err,
		}
	}

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
