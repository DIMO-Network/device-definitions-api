package commands

import (
	"context"
	"database/sql"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type CreateDeviceDefinitionCommand struct {
	Source           string                                  `json:"source"`
	Make             string                                  `json:"make"`
	Model            string                                  `json:"model"`
	Year             int                                     `json:"year"`
	DeviceTypeID     string                                  `json:"device_type_id"`
	DeviceAttributes []*UpdateDeviceDefinitionAttributeModel `json:"deviceAttributes"`
}

type UpdateDeviceDefinitionAttributeModel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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

	const (
		deviceTypeVehicle = "vehicle"
	)

	if len(command.DeviceTypeID) == 0 {
		command.DeviceTypeID = deviceTypeVehicle
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
	deviceTypeInfo := make(map[string]interface{})
	metaData := make(map[string]interface{})
	var ai map[string][]coremodels.GetDeviceTypeAttributeQueryResult
	if err := dt.Properties.Unmarshal(&ai); err == nil {
		filterProperty := func(name string, items []coremodels.GetDeviceTypeAttributeQueryResult) *coremodels.GetDeviceTypeAttributeQueryResult {
			for _, attribute := range items {
				if name == attribute.Name {
					return &attribute
				}
			}
			return nil
		}

		for _, prop := range command.DeviceAttributes {
			property := filterProperty(prop.Name, ai["properties"])

			if property == nil {
				return nil, &exceptions.ValidationError{
					Err: fmt.Errorf("invalid property %s", prop.Name),
				}
			}

			if property.Required && len(prop.Value) == 0 {
				return nil, &exceptions.ValidationError{
					Err: fmt.Errorf("property %s is required", prop.Name),
				}
			}

			metaData[property.Name] = prop.Value
		}
	}

	deviceTypeInfo[dt.Metadatakey] = metaData

	dd, err := ch.Repository.GetOrCreate(ctx, command.Source, command.Make, command.Model, command.Year, command.DeviceTypeID, deviceTypeInfo)

	if err != nil {
		return nil, err
	}

	return CreateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}
