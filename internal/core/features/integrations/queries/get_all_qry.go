package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetAllQuery struct {
}

type GetAllQueryResult struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Style  string `json:"style"`
	Vendor string `json:"vendor"`
}

func (*GetAllQuery) Key() string { return "GetAllQuery" }

type GetAllQueryHandler struct {
	Db *sql.DB
}

func NewGetAllQueryHandler(db *sql.DB) GetAllQueryHandler {
	return GetAllQueryHandler{
		Db: db,
	}
}

func (ch GetAllQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	all, err := models.Integrations(qm.Limit(100)).All(ctx, ch.Db)
	if err != nil {
		return nil, &common.InternalError{
			Err: fmt.Errorf("failed to get integrations"),
		}
	}

	var result []GetAllQueryResult
	for _, v := range all {
		result = append(result, GetAllQueryResult{
			ID:     v.ID,
			Type:   v.Type,
			Style:  v.Style,
			Vendor: v.Vendor,
		})
	}

	return result, nil

}
