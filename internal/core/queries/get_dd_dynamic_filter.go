//nolint:tagliatelle
package queries

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/null/v8"
)

type GetDeviceDefinitionByDynamicFilterQuery struct {
	DefinitionID    string   `json:"definition_id"`
	Year            int      `json:"year"`
	Model           string   `json:"model"`
	VerifiedVinList []string `json:"verified_vin_list"`
	PageIndex       int      `json:"page_index"`
	PageSize        int      `json:"page_size"`
	MakeSlug        string
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
	DBS        func() *db.ReaderWriter
	onChainSvc gateways.DeviceDefinitionOnChainService
}

func NewGetDeviceDefinitionByDynamicFilterQueryHandler(dbs func() *db.ReaderWriter, onChainSvc gateways.DeviceDefinitionOnChainService) GetDeviceDefinitionByDynamicFilterQueryHandler {
	return GetDeviceDefinitionByDynamicFilterQueryHandler{
		DBS:        dbs,
		onChainSvc: onChainSvc,
	}
}

func (ch GetDeviceDefinitionByDynamicFilterQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByDynamicFilterQuery)

	if len(qry.DefinitionID) > 1 {
		dd, _, err := ch.onChainSvc.GetDefinitionByID(ctx, qry.DefinitionID, ch.DBS().Reader)
		if err != nil {
			return nil, err
		}
		dds := make([]DeviceDefinitionQueryResponse, 1)
		dds[0] = ch.buildDeviceDefinitionQueryResponse(ctx, dd)
		return dds, nil
	}

	manufacturerID := types.NullDecimal{}

	if len(qry.MakeSlug) > 1 {
		manufacturer, err := ch.onChainSvc.GetManufacturer(ctx, qry.MakeSlug, ch.DBS().Reader)
		if err != nil {
			return nil, err
		}
		manufacturerID = types.NewNullDecimal(decimal.New(int64(manufacturer.TokenID), 0))
	}

	definitions, err := ch.onChainSvc.GetDeviceDefinitions(ctx, manufacturerID, "", qry.Model, qry.Year, int32(qry.PageIndex), int32(qry.PageSize))
	if err != nil {
		return nil, err
	}

	dd := make([]DeviceDefinitionQueryResponse, len(definitions))
	for i, item := range definitions {
		dd[i] = ch.buildDeviceDefinitionQueryResponse(ctx, &item)
	}

	return dd, err

}

func (ch GetDeviceDefinitionByDynamicFilterQueryHandler) buildDeviceDefinitionQueryResponse(ctx context.Context, dd *models.DeviceDefinitionTablelandModel) DeviceDefinitionQueryResponse {
	if dd == nil {
		return DeviceDefinitionQueryResponse{}
	}
	split := strings.Split(dd.ID, "_")
	manufacturerSlug := split[0]
	manufacturer, _ := ch.onChainSvc.GetManufacturer(ctx, manufacturerSlug, ch.DBS().Reader)
	mdStr := []byte("{}")
	if dd.Metadata != nil {
		mdStr, _ = json.Marshal(dd.Metadata)
	}

	return DeviceDefinitionQueryResponse{
		ID:           dd.ID,
		NameSlug:     dd.ID,
		Model:        dd.Model,
		Year:         dd.Year,
		ImageURL:     null.StringFrom(dd.ImageURI),
		Verified:     true,
		DeviceMakeID: strconv.Itoa(manufacturer.TokenID),
		Make:         manufacturer.Name,
		Metadata:     null.JSONFrom(mdStr),
	}
}
