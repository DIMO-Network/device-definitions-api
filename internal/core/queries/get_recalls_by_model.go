package queries

import (
	"context"
	"database/sql"
	"fmt"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetRecallsByModelQuery struct {
	Model string `json:"model"`
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

	dds, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Model.EQ(qry.Model),
		models.DeviceDefinitionWhere.Year.GTE(cutoffYear), qm.Select(models.DeviceDefinitionColumns.ID)).All(ctx, qh.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []p_grpc.RecallItem{}, nil
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
			return []coremodels.GetIntegrationQueryResult{}, nil
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
			Description:        "", // todo: ask james.
			//Date:               v.DataRcdate.UnixMilli(),
			Year: int32(v.DataYeartxt),
		})
	}

	return result, nil
}
