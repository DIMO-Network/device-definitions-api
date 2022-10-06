package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type CreateDeviceMakeCommand struct {
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
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

	dm, err := ch.Repository.GetOrCreate(ctx, command.Name, command.LogoURL)

	if err != nil {
		return nil, err
	}

	return CreateDeviceMakeCommandResult{ID: dm.ID}, nil
}
