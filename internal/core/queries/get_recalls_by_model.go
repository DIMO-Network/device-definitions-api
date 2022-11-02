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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	all, err := models.DeviceNhtsaRecalls(models.DeviceNhtsaRecallWhere.DeviceDefinitionID.EQ(null.StringFrom(qry.DeviceDefinitionID)),
		qm.Load(models.DeviceNhtsaRecallRels.DeviceDefinition),
		qm.Load(qm.Rels(models.DeviceNhtsaRecallRels.DeviceDefinition, models.DeviceDefinitionRels.DeviceMake))).
		All(ctx, qh.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &p_grpc.GetRecallsResponse{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get nthsa"),
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
			ManufactureCampNo:  v.DataMfgcampno,
			ConsequenceDefect:  v.DataConequenceDefect,
		})
	}

	return result, nil
}
