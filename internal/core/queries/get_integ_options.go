package queries

import (
	"context"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

type GetIntegrationOptionsQuery struct {
	MakeID string
}

func (*GetIntegrationOptionsQuery) Key() string {
	return "GetIntegrationOptionsQuery"
}

type GetIntegrationOptionsQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetIntegrationOptionsQueryHandler(dbs func() *db.ReaderWriter) GetIntegrationOptionsQueryHandler {
	return GetIntegrationOptionsQueryHandler{DBS: dbs}
}

func (ch GetIntegrationOptionsQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetIntegrationOptionsQuery)
	// could pull this one from cache since doesn't change often

	// raw query

	// build up grpc object
	return qry, nil
}
