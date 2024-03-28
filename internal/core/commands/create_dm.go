//nolint:tagliatelle
package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
)

type CreateDeviceMakeCommand struct {
	Name               string `json:"name"`
	LogoURL            string `json:"logo_url"`
	ExternalIDs        string `json:"external_ids,omitempty"`
	Metadata           string `json:"metadata,omitempty"`
	HardwareTemplateID string `json:"hardware_template_id,omitempty"`
}

type CreateDeviceMakeCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceMakeCommand) Key() string { return "CreateDeviceMakeCommand" }

type CreateDeviceMakeCommandHandler struct {
	Repository repositories.DeviceMakeRepository
}

func NewCreateDeviceMakeCommandHandler(repository repositories.DeviceMakeRepository) CreateDeviceMakeCommandHandler {
	return CreateDeviceMakeCommandHandler{Repository: repository}
}

func (ch CreateDeviceMakeCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateDeviceMakeCommand)

	dm, err := ch.Repository.GetOrCreate(ctx, command.Name, command.LogoURL, command.ExternalIDs, command.Metadata, command.HardwareTemplateID)

	if err != nil {
		return nil, err
	}

	return CreateDeviceMakeCommandResult{ID: dm.ID}, nil
}
