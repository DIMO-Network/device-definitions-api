package commands

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/shared/db"
)

type DemoCommand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *DemoCommand) Key() string {
	return "DemoCommand"
}

type DemoCommandResult struct {
	ID string `json:"id"`
}

type DemoCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewDemoCommandHandler(dbs func() *db.ReaderWriter) DemoCommandHandler {
	return DemoCommandHandler{DBS: dbs}
}

func (c DemoCommandHandler) Handle(ctx context.Context, cmd mediator.Message) (interface{}, error) {
	command := cmd.(*DemoCommand)
	fmt.Printf("DemoCommandHandler handling the command with name: %s\n", command.Name)

	return DemoCommandResult{ID: "test"}, nil
}
