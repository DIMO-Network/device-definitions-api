package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UpdateIntegrationFeatureCommand struct {
	ID              string  `json:"id"`
	ElasticProperty string  `json:"elastic_property"`
	DisplayName     string  `json:"display_name"`
	CSSIcon         string  `json:"css_icon"`
	FeatureWeight   float64 `json:"feature_weight"`
}

type UpdateIntegrationFeatureResult struct {
	ID string `json:"id"`
}

func (*UpdateIntegrationFeatureCommand) Key() string { return "UpdateIntegrationFeatureCommand" }

type UpdateIntegrationFeatureCommandHandler struct {
	DBS func() *db.ReaderWriter
}

func NewUpdateIntegrationFeatureCommandHandler(dbs func() *db.ReaderWriter) UpdateIntegrationFeatureCommandHandler {
	return UpdateIntegrationFeatureCommandHandler{DBS: dbs}
}

func (ch UpdateIntegrationFeatureCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateIntegrationFeatureCommand)

	feature, err := models.IntegrationFeatures(models.IntegrationFeatureWhere.FeatureKey.EQ(command.ID)).One(ctx, ch.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: fmt.Errorf("could not find integration feature id: %s", command.ID),
			}
		}
	}

	feature.ElasticProperty = command.ElasticProperty
	feature.DisplayName = command.DisplayName
	feature.CSSIcon = null.StringFrom(command.CSSIcon)
	feature.FeatureWeight = null.Float64From(command.FeatureWeight)

	if _, err := feature.Update(ctx, ch.DBS().Writer.DB, boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return UpdateIntegrationFeatureResult{ID: feature.FeatureKey}, nil
}
