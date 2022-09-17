package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const inner_join_query_format = "%s on %s.%s = %s.%s"
const and_eq_query_format = "%s = %s"
const and_like_query_format = "%s ilike %s"

type GetDeviceDefinitionByDynamicFilterQuery struct {
	MakeID             string `json:"make_id"`
	IntegrationID      string `json:"integration_id"`
	DeviceDefinitionID string `json:"device_definition_id"`
	Year               int    `json:"year"`
	Model              string `json:"model"`
	VerifiedVin        bool   `json:"verified_vin"`
	PageIndex          int    `json:"page_index"`
	PageSize           int    `json:"page_size"`
}

func (*GetDeviceDefinitionByDynamicFilterQuery) Key() string {
	return "GetDeviceDefinitionByDynamicFilterQuery"
}

type GetDeviceDefinitionByDynamicFilterQueryHandler struct {
	DBS func() *db.ReaderWriter
}

func NewGetDeviceDefinitionByDynamicFilterQueryHandler(dbs func() *db.ReaderWriter) GetDeviceDefinitionByDynamicFilterQueryHandler {
	return GetDeviceDefinitionByDynamicFilterQueryHandler{
		DBS: dbs,
	}
}

func (ch GetDeviceDefinitionByDynamicFilterQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByDynamicFilterQuery)

	var queryMods []qm.QueryMod

	if len(qry.DeviceDefinitionID) > 1 {
		queryMods = append(queryMods, models.DeviceDefinitionWhere.ID.EQ(string(qry.DeviceDefinitionID)))
	}

	if len(qry.IntegrationID) > 1 {
		queryMods = append(queryMods,
			qm.InnerJoin(fmt.Sprintf(inner_join_query_format, models.TableNames.DeviceIntegrations,
				models.TableNames.DeviceIntegrations,
				models.DeviceIntegrationColumns.DeviceDefinitionID,
				models.TableNames.DeviceDefinitions,
				models.DeviceDefinitionColumns.ID),
			),
			qm.And(fmt.Sprintf(and_eq_query_format, models.DeviceIntegrationColumns.IntegrationID, qry.IntegrationID)))
	}

	if len(qry.MakeID) > 1 {
		queryMods = append(queryMods, qm.And(fmt.Sprintf(and_eq_query_format, models.DeviceDefinitionColumns.DeviceMakeID, qry.MakeID)))
	}

	if qry.Year > 1980 && qry.Year < 2999 {
		queryMods = append(queryMods, models.DeviceDefinitionWhere.Year.EQ(int16(qry.Year)))
	}

	if len(qry.Model) > 1 {
		queryMods = append(queryMods, qm.And(fmt.Sprintf(and_like_query_format, models.DeviceDefinitionColumns.Model, qry.Model+"%")))
	}

	queryMods = append(queryMods,
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.OrderBy(fmt.Sprintf("%s, %s, %s",
			models.DeviceDefinitionColumns.DeviceMakeID,
			models.DeviceDefinitionColumns.Year,
			models.DeviceDefinitionColumns.Model)),
		qm.Limit(qry.PageSize),
		qm.Offset(qry.PageIndex*qry.PageSize))

	dd, err := models.DeviceDefinitions(queryMods...).All(ctx, ch.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.DeviceDefinition{}, err
		}

		return nil, &exceptions.InternalError{Err: err}
	}

	return dd, err

}
