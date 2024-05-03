//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const (
	innerJoinQueryFormat = "%s on %s.%s = %s.%s"
	andEqQueryFormat     = "%s = '%s'"
	andLikeQueryFormat   = "%s ilike '%s'"
)

type GetDeviceDefinitionByDynamicFilterQuery struct {
	MakeID             string   `json:"make_id"`
	IntegrationID      string   `json:"integration_id"`
	DeviceDefinitionID string   `json:"device_definition_id"`
	Year               int      `json:"year"`
	Model              string   `json:"model"`
	VerifiedVinList    []string `json:"verified_vin_list"`
	PageIndex          int      `json:"page_index"`
	PageSize           int      `json:"page_size"`
}

type DeviceDefinitionQueryResponse struct {
	ID           string      `json:"id"`
	NameSlug     string      `json:"name_slug"`
	Model        string      `json:"model"`
	Year         int         `json:"year"`
	ImageURL     null.String `json:"image_url,omitempty"`
	CreatedAt    time.Time   `json:"created_at,omitempty"`
	UpdatedAt    time.Time   `json:"updated_at,omitempty"`
	Metadata     null.JSON   `json:"metadata"`
	Source       null.String `json:"source"`
	Verified     bool        `json:"verified"`
	ExternalID   null.String `json:"external_id"`
	DeviceMakeID string      `json:"device_make_id"`
	Make         string      `json:"make"`
	ExternalIDs  null.JSON   `json:"external_ids"`
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

	if len(qry.VerifiedVinList) > 1 {
		queryMods = append(queryMods, models.DeviceDefinitionWhere.ID.IN(qry.VerifiedVinList))
	}

	if len(qry.IntegrationID) > 1 {
		queryMods = append(queryMods,
			qm.InnerJoin(fmt.Sprintf(innerJoinQueryFormat,
				models.TableNames.DeviceIntegrations,
				models.TableNames.DeviceIntegrations,
				models.DeviceIntegrationColumns.DeviceDefinitionID,
				models.TableNames.DeviceDefinitions,
				models.DeviceDefinitionColumns.ID),
			),
			qm.And(fmt.Sprintf(andEqQueryFormat, models.DeviceIntegrationColumns.IntegrationID, qry.IntegrationID)))
	}

	if len(qry.MakeID) > 1 {
		queryMods = append(queryMods, qm.And(fmt.Sprintf(andEqQueryFormat, models.DeviceDefinitionColumns.DeviceMakeID, qry.MakeID)))
	}

	if qry.Year > 1980 && qry.Year < 2999 {
		queryMods = append(queryMods, models.DeviceDefinitionWhere.Year.EQ(int16(qry.Year)))
	}

	if len(qry.Model) > 1 {
		queryMods = append(queryMods, qm.And(fmt.Sprintf(andLikeQueryFormat, models.DeviceDefinitionColumns.Model, qry.Model+"%")))
	}

	queryMods = append(queryMods,
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.OrderBy(fmt.Sprintf("%s, %s, %s",
			models.DeviceDefinitionColumns.DeviceMakeID,
			models.DeviceDefinitionColumns.Year,
			models.DeviceDefinitionColumns.Model)),
		qm.Limit(qry.PageSize),
		qm.Offset(qry.PageIndex*qry.PageSize))

	all, err := models.DeviceDefinitions(queryMods...).All(ctx, ch.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.DeviceDefinition{}, err
		}

		return nil, &exceptions.InternalError{Err: err}
	}

	dd := make([]DeviceDefinitionQueryResponse, len(all))
	for i, item := range all {
		dd[i] = buildDeviceDefinitionQueryResponse(item)
	}

	return dd, err

}

func buildDeviceDefinitionQueryResponse(dd *models.DeviceDefinition) DeviceDefinitionQueryResponse {

	if dd == nil {
		return DeviceDefinitionQueryResponse{}
	}

	return DeviceDefinitionQueryResponse{
		ID:           dd.ID,
		NameSlug:     dd.NameSlug,
		Model:        dd.Model,
		Year:         int(dd.Year),
		ImageURL:     null.StringFrom(common.GetDefaultImageURL(dd)),
		CreatedAt:    dd.CreatedAt,
		UpdatedAt:    dd.UpdatedAt,
		Source:       dd.Source,
		Verified:     dd.Verified,
		ExternalID:   dd.ExternalID,
		DeviceMakeID: dd.DeviceMakeID,
		Make:         dd.R.DeviceMake.Name,
		Metadata:     dd.Metadata,
		ExternalIDs:  dd.ExternalIds,
	}
}
