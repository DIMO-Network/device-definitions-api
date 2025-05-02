//nolint:tagliatelle
package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
)

type CreateDeviceStyleCommand struct {
	DefinitionID       string `json:"definition_id"`
	Name               string `json:"name"`
	ExternalStyleID    string `json:"external_style_id"`
	Source             string `json:"source"`
	SubModel           string `json:"sub_model"`
	HardwareTemplateID string `json:"hardware_template_id"`
}

type CreateDeviceStyleCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceStyleCommand) Key() string { return "CreateDeviceStyleCommand" }

type CreateDeviceStyleCommandHandler struct {
	repository repositories.DeviceStyleRepository
}

func NewCreateDeviceStyleCommandHandler(repository repositories.DeviceStyleRepository) CreateDeviceStyleCommandHandler {
	return CreateDeviceStyleCommandHandler{repository: repository}
}

func (ch CreateDeviceStyleCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateDeviceStyleCommand)

	ds, err := ch.repository.Create(ctx, command.DefinitionID, command.Name, command.ExternalStyleID, command.Source, command.SubModel, command.HardwareTemplateID)

	if err != nil {
		return nil, err
	}

	return CreateDeviceStyleCommandResult{ID: ds.ID}, nil
}
