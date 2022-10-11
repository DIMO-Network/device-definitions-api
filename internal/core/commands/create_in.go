package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/TheFellow/go-mediator/mediator"
)

type CreateIntegrationCommand struct {
	Vendor string `json:"vendor"`
	Type   string `json:"type"`
	Style  string `json:"style"`
}

type CreateIntegrationCommandResult struct {
	ID string `json:"id"`
}

func (*CreateIntegrationCommand) Key() string { return "CreateIntegrationCommand" }

type CreateIntegrationCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewCreateIntegrationCommandHandler(dbs func() *db.ReaderWriter) CreateIntegrationCommandHandler {
	return CreateIntegrationCommandHandler{DBS: dbs}
}

func (ch CreateIntegrationCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateIntegrationCommand)

	i := models.Integration{}
	i.ID = ksuid.New().String()
	i.Vendor = command.Vendor
	i.Type = command.Type
	i.Style = command.Style
	err := i.Insert(ctx, ch.DBS().Writer, boil.Infer())

	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrapf(err, "error inserting integration: %s", command.Vendor)}
	}

	return CreateIntegrationCommandResult{ID: i.ID}, nil
}
