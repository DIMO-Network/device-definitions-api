package queries

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"

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

	all, err := models.DeviceNhtsaRecalls(models.DeviceNhtsaRecallWhere.DeviceDefinitionID.IN(ddIDs),
		qm.Load(models.DeviceNhtsaRecallRels.DeviceDefinition),
		qm.Load(qm.Rels(models.DeviceNhtsaRecallRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake))).
		All(ctx, qh.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &p_grpc.GetRecallsResponse{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get Nhtsa"),
		}
	}

	result := &p_grpc.GetRecallsResponse{}

	for _, v := range all {
		result.Recalls = append(result.Recalls, &p_grpc.RecallItem{
			DeviceDefinitionId: v.DeviceDefinitionID.String,
			Name:               fmt.Sprintf("%d %s %s", v.R.DeviceDefinition.Year, v.R.DeviceDefinition.R.DeviceMake.Name, v.R.DeviceDefinition.Model),
			Description:        v.DataDescDefect,
			Date:               timestamppb.New(v.DataRcdate),
			Year:               int32(v.DataYeartxt),
			ComponentName:      v.DataCompname,
		})
	}

	return result, nil
}
