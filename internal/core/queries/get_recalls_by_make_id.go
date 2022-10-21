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
	// todo what to return?
	qry := query.(*GetRecallsByMakeQuery)

	dds, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(qry.MakeID),
		models.DeviceDefinitionWhere.Year.GTE(cutoffYear), qm.Select(models.DeviceDefinitionColumns.ID)).All(ctx, qh.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []coremodels.GetIntegrationQueryResult{}, nil
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integrations"),
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

	// todo: return array of nhtsa recalls with properties same as proto

	result := make([]coremodels.GetIntegrationQueryResult, len(all))
	for i, v := range all {
		im := new(coremodels.IntegrationsMetadata)
		if v.Metadata.Valid {
			err = v.Metadata.Unmarshal(&im)

			if err != nil {
				return nil, &exceptions.InternalError{
					Err: fmt.Errorf("failed to unmarshall integration metadata id %s", v.ID),
				}
			}
		}
		result[i] = p_grpc.RecallItem{ // se puede? only for grpc no rest
			ID:                      v.ID,
			Type:                    v.Type,
			Style:                   v.Style,
			Vendor:                  v.Vendor,
			AutoPiDefaultTemplateID: im.AutoPiDefaultTemplateID,
			RefreshLimitSecs:        v.RefreshLimitSecs,
		}
		if im.AutoPiPowertrainToTemplateID != nil {
			result[i].AutoPiPowertrainToTemplateID = im.AutoPiPowertrainToTemplateID
		}
	}

	return result, nil
}
