package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type CreateDeviceStyleCommand struct {
	DeviceDefinitionID string `json:"device_definition_id"`
	Name               string `json:"name"`
	ExternalStyleID    string `json:"external_style_id"`
	Source             string `json:"source"`
	SubModel           string `json:"sub_model"`
}

type CreateDeviceStyleCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceStyleCommand) Key() string { return "CreateDeviceStyleCommand" }

type CreateDeviceStyleCommandHandler struct {
	Repository repositories.DeviceStyleRepository
}

func NewCreateDeviceStyleCommandHandler(repository repositories.DeviceStyleRepository) CreateDeviceStyleCommandHandler {
	return CreateDeviceStyleCommandHandler{Repository: repository}
}

func (ch CreateDeviceStyleCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateDeviceStyleCommand)

	ds, err := ch.Repository.Create(ctx, command.DeviceDefinitionID, command.Name, command.ExternalStyleID, command.Source, command.SubModel)

	if err != nil {
		return nil, err
	}

	return CreateDeviceStyleCommandResult{ID: ds.ID}, nil
}