package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type CreateIntegrationFeatureCommand struct {
	ID              string  `json:"id"`
	ElasticProperty string  `json:"elastic_property"`
	DisplayName     string  `json:"display_name"`
	CSSIcon         string  `json:"css_icon"`
	FeatureWeight   float64 `json:"feature_weight"`
}

type CreateIntegrationFeatureCommandResult struct {
	ID string `json:"id"`
}

func (*CreateIntegrationFeatureCommand) Key() string { return "CreateIntegrationFeatureCommand" }

type CreateIntegrationFeatureCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewCreateIntegrationFeatureCommandHandler(dbs func() *db.ReaderWriter) CreateIntegrationFeatureCommandHandler {
	return CreateIntegrationFeatureCommandHandler{DBS: dbs}
}

func (ch CreateIntegrationFeatureCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*CreateIntegrationFeatureCommand)

	feature := models.IntegrationFeature{}
	feature.FeatureKey = command.ID
	feature.ElasticProperty = command.ElasticProperty
	feature.DisplayName = command.DisplayName
	feature.CSSIcon = null.StringFrom(command.CSSIcon)
	feature.FeatureWeight = null.Float64From(command.FeatureWeight)

	err := feature.Insert(ctx, ch.DBS().Writer, boil.Infer())

	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrapf(err, "error inserting integration feature: %s", feature.FeatureKey)}
	}

	return CreateIntegrationFeatureCommandResult{ID: feature.FeatureKey}, nil
}
