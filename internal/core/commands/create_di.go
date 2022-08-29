package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type CreateDeviceIntegrationCommand struct {
	DeviceDefinitionID string `json:"device_definition_id"`
	IntegrationID      string `json:"integration_id"`
	Region             string `json:"region"`
}

type CreateDeviceIntegrationCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceIntegrationCommand) Key() string { return "CreateDeviceIntegrationCommand" }

type CreateDeviceIntegrationCommandHandler struct {
	Repository repositories.DeviceIntegrationRepository
}

func NewCreateDeviceIntegrationCommandHandler(repository repositories.DeviceIntegrationRepository) CreateDeviceIntegrationCommandHandler {
	return CreateDeviceIntegrationCommandHandler{Repository: repository}
}

func (ch CreateDeviceIntegrationCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateDeviceIntegrationCommand)

	di, err := ch.Repository.Create(ctx, command.DeviceDefinitionID, command.IntegrationID, command.Region)

	if err != nil {
		return nil, err
	}

	return CreateDeviceIntegrationCommandResult{ID: di.IntegrationID}, nil
}
