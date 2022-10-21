package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetRecallsByMakeQuery struct {
	MakeID string `json:"make_id"`
}

func (*GetRecallsByMakeQuery) Key() string { return "GetRecallsByMakeQuery" }

type GetRecallsByMakeQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetRecallsByMakeQueryHandler(dbs func() *db.ReaderWriter) GetRecallsByMakeQueryHandler {
	return GetRecallsByMakeQueryHandler{DBS: dbs}
}

const cutoffYear = 2005

func (qh GetRecallsByMakeQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetRecallsByMakeQuery)

	dds, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(qry.MakeID),
		models.DeviceDefinitionWhere.Year.GTE(cutoffYear), qm.Select(models.DeviceDefinitionColumns.ID)).All(ctx, qh.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p_grpc.GetRecallsResponse{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device definitions"),
		}
	}

	ddIDs := make([]string, len(dds))
	for i, dd := range dds {
		ddIDs[i] = dd.ID
	}

	// todo not sure if in ddIDs will work
	all, err := models.DeviceNhtsaRecalls(qm.AndIn("device_definition_id in ?", ddIDs)).All(ctx, qh.DBS().Reader)
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
