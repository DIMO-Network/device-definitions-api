package commands

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type CreateDeviceDefinitionCommand struct {
	Source string `json:"source"`
	Make   string `json:"make"`
	Model  string `json:"model"`
	Year   int    `json:"year"`
}

type CreateDeviceDefinitionCommandResult struct {
	Id string `json:"id"`
}

func (*CreateDeviceDefinitionCommand) Key() string { return "CreateDeviceDefinitionCommand" }

type CreateDeviceDefinitionCommandHandler struct {
	Repository repositories.DeviceDefinitionRepository
}

func NewCreateDeviceDefinitionCommandHandler(repository repositories.DeviceDefinitionRepository) CreateDeviceDefinitionCommandHandler {
	return CreateDeviceDefinitionCommandHandler{Repository: repository}
}

func (ch CreateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateDeviceDefinitionCommand)

	dd, err := ch.Repository.GetOrCreate(ctx, command.Source, command.Make, command.Model, command.Year)

	if err != nil {
		return nil, err
	}

	return CreateDeviceDefinitionCommandResult{Id: dd.ID}, nil
}
