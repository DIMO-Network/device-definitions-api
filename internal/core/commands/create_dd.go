package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type CreateDeviceDefinitionCommand struct {
	Source string `json:"source"`
	Make   string `json:"make"`
	Model  string `json:"model"`
	Year   int    `json:"year"`
	// DeviceTypeID comes from the device_types.id table, determines what kind of device this is, typically a vehicle
	DeviceTypeID string `json:"device_type_id"`
	// DeviceAttributes sets definition metadata eg. vehicle info. Allowed key/values are defined in device_types.properties
	DeviceAttributes []*coremodels.UpdateDeviceTypeAttribute `json:"deviceAttributes"`
}

type CreateDeviceDefinitionCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceDefinitionCommand) Key() string { return "CreateDeviceDefinitionCommand" }

type CreateDeviceDefinitionCommandHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
}

func NewCreateDeviceDefinitionCommandHandler(repository repositories.DeviceDefinitionRepository, dbs func() *db.ReaderWriter) CreateDeviceDefinitionCommandHandler {
	return CreateDeviceDefinitionCommandHandler{Repository: repository, DBS: dbs}
}

func (ch CreateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	command := query.(*CreateDeviceDefinitionCommand)

	if len(command.DeviceTypeID) == 0 {
		command.DeviceTypeID = common.DefaultDeviceType
	}

	// Resolve attributes by device types
	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.DeviceTypeID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to get device types"),
			}
		}
	}

	// attribute info
	deviceTypeInfo, err := common.BuildDeviceTypeAttributes(command.DeviceAttributes, dt)
	if err != nil {
		return nil, err
	}

	dd, err := ch.Repository.GetOrCreate(ctx, command.Source, command.Make, command.Model, command.Year, command.DeviceTypeID, deviceTypeInfo)
	if err != nil {
		return nil, err
	}

	return CreateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
