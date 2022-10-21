package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/volatiletech/null/v8"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
)

type GetRecallsByModelQuery struct {
	DeviceDefinitionID string `json:"deviceDefinitionID"`
}

func (*GetRecallsByModelQuery) Key() string { return "GetRecallsByModelQuery" }

type GetRecallsByModelQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetRecallsByModelQueryHandler(dbs func() *db.ReaderWriter) GetRecallsByModelQueryHandler {
	return GetRecallsByModelQueryHandler{DBS: dbs}
}

func (qh GetRecallsByModelQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetRecallsByModelQuery)

	all, err := models.DeviceNhtsaRecalls(models.DeviceNhtsaRecallWhere.DeviceDefinitionID.EQ(null.StringFrom(qry.DeviceDefinitionID))).
		All(ctx, qh.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &p_grpc.GetRecallsResponse{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integrations"),
		}
	}

	result := &p_grpc.GetRecallsResponse{}

	for _, v := range all {
		result.Recalls = append(result.Recalls, &p_grpc.RecallItem{
			DeviceDefinitionId: v.DeviceDefinitionID.String,
			Name:               fmt.Sprintf("%d %s %s", v.R.DeviceDefinition.Year, v.R.DeviceDefinition.R.DeviceMake.Name, v.R.DeviceDefinition.Model),
			Description:        v.DataDescDefect,
			//Date:               v.DataRcdate.UnixMilli(),
			Year: int32(v.DataYeartxt),
		})
	}

	return result, nil
}
