package commands

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/volatiletech/null/v8"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/pkg/errors"
)

type CreateDeviceDefinitionCommand struct {
	Source             string `json:"source"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	Year               int    `json:"year"`
	HardwareTemplateID string `json:"hardware_template_id,omitempty"`
	// DeviceTypeID comes from the device_types.id table, determines what kind of device this is, typically a vehicle
	DeviceTypeID string `json:"device_type_id"`
	// DeviceAttributes sets definition metadata eg. vehicle info. Allowed key/values are defined in device_types.properties
	DeviceAttributes []*coremodels.UpdateDeviceTypeAttribute `json:"deviceAttributes"`
	Verified         bool                                    `json:"verified"`
}

type CreateDeviceDefinitionCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceDefinitionCommand) Key() string { return "CreateDeviceDefinitionCommand" }

type CreateDeviceDefinitionCommandHandler struct {
	Repository            repositories.DeviceDefinitionRepository
	DBS                   func() *db.ReaderWriter
	powerTrainTypeService services.PowerTrainTypeService
}

func NewCreateDeviceDefinitionCommandHandler(repository repositories.DeviceDefinitionRepository, dbs func() *db.ReaderWriter, powerTrainTypeService services.PowerTrainTypeService) CreateDeviceDefinitionCommandHandler {
	return CreateDeviceDefinitionCommandHandler{Repository: repository, DBS: dbs, powerTrainTypeService: powerTrainTypeService}
}

func (ch CreateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	command := query.(*CreateDeviceDefinitionCommand)

	if len(command.DeviceTypeID) == 0 {
		command.DeviceTypeID = common.DefaultDeviceType
	}

	// Resolve attributes by device types
	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.DeviceTypeID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.ValidationError{
				Err: fmt.Errorf("device type id: %s not found when updating a definition", command.DeviceTypeID),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device types"),
		}
	}

	// Validate if powertraintype exists
	powerTrainExists := false
	if len(command.DeviceAttributes) > 0 {
		for _, item := range command.DeviceAttributes {
			if item.Name == common.PowerTrainType && len(item.Value) > 0 {
				powerTrainExists = true
				break
			}
		}
	}
	if !powerTrainExists {
		powerTrainTypeValue, err := ch.powerTrainTypeService.ResolvePowerTrainType(ctx, common.SlugString(command.Make), common.SlugString(command.Model), nil, null.JSON{}, null.JSON{})
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to get powertraintype"),
			}
		}
		if powerTrainTypeValue != "" {
			command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
				Name:  common.PowerTrainType,
				Value: powerTrainTypeValue,
			})
		}
	}
	// attribute info
	deviceTypeInfo, err := common.BuildDeviceTypeAttributes(command.DeviceAttributes, dt)
	if err != nil {
		return nil, err
	}

	dd, err := ch.Repository.GetOrCreate(ctx, nil, command.Source, "", command.Make, command.Model, command.Year, command.DeviceTypeID, deviceTypeInfo, command.Verified, &command.HardwareTemplateID)
	if err != nil {
		return nil, err
	}

	return CreateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
