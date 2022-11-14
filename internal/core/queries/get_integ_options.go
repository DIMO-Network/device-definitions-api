package queries

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries"
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
	// could pull this one from cache since doesn't change often, although we have frontend caching from explorer
	var options []*IntegrationOption
	// raw query to keep performant
	err := queries.Raw(`select di.integration_id, i.vendor, di.region from device_integrations di
	join integrations i on i.id = di.integration_id
	join device_definitions dd on dd.id = di.device_definition_id
	where dd.device_make_id = $1
	group by di.integration_id, i.vendor, di.region`, qry.MakeID).Bind(ctx, ch.DBS().Reader, &options)

	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration options for make %s", qry.MakeID),
		}
	}

	// build up grpc object
	resp := p_grpc.GetIntegrationOptionsResponse{IntegrationOptions: make([]*p_grpc.GetIntegrationOptionItem, len(options))}
	for i, option := range options {
		resp.IntegrationOptions[i] = &p_grpc.GetIntegrationOptionItem{
			IntegrationId:     option.IntegrationID,
			IntegrationVendor: option.IntegrationVendor,
			Region:            option.Region,
		}
	}

	return resp, nil
}

type IntegrationOption struct {
	IntegrationID     string `boil:"integration_id"`
	IntegrationVendor string `boil:"vendor"`
	Region            string `boil:"region"`
}
