package queries

import (
	"context"
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceDefinitionBySlugQuery struct {
	DefinitionID string `json:"definitionId"`
}

func (*GetDeviceDefinitionBySlugQuery) Key() string { return "GetDeviceDefinitionBySlugQuery" }

type GetDeviceDefinitionBySlugQueryHandler struct {
	definitionsOnChainService gateways.DeviceDefinitionOnChainService
	dbs                       func() *db.ReaderWriter
}

func NewGetDeviceDefinitionBySlugQueryHandler(ddOnChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetDeviceDefinitionBySlugQueryHandler {
	return GetDeviceDefinitionBySlugQueryHandler{
		definitionsOnChainService: ddOnChainSvc,
		dbs:                       dbs,
	}
}

// Handle the request - pretty much only used by grpc response
func (ch GetDeviceDefinitionBySlugQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionBySlugQuery)

	dd, manufacturerID, err := ch.definitionsOnChainService.GetDefinitionByID(ctx, qry.DefinitionID, ch.dbs().Reader)

	if err != nil {
		return nil, errors.Wrap(err, "handler failed to get dd by slug "+qry.DefinitionID)
	}

	if dd == nil {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find definition id: %s", qry.DefinitionID),
		}
	}

	dbDefinition, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.NameSlug.EQ(dd.ID),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
	).One(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}

	result := BuildFromDeviceDefinitionToGRPCResult(dbDefinition, dd, manufacturerID)

	return result, nil
}

func BuildFromDeviceDefinitionToGRPCResult(dd *models.DeviceDefinition, tbl *gateways.DeviceDefinitionTablelandModel, manufacturerID *big.Int) *grpc.GetDeviceDefinitionItemResponse {
	rp := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: tbl.KSUID,
		Ksuid:              tbl.KSUID,
		Model:              tbl.Model,
		Year:               int32(tbl.Year),
		Id:                 tbl.ID,
		NameSlug:           tbl.ID, //nolint
		Name:               common.BuildDeviceDefinitionName(int16(tbl.Year), dd.R.DeviceMake.Name, tbl.Model),
		ImageUrl:           tbl.ImageURI,

		HardwareTemplateId: "130", //used for the autopi template id, which should always be 130 now
		Make: &grpc.DeviceMake{
			Id:              dd.R.DeviceMake.ID,
			Name:            dd.R.DeviceMake.Name,
			LogoUrl:         dd.R.DeviceMake.LogoURL.String,
			OemPlatformName: dd.R.DeviceMake.OemPlatformName.String,
			NameSlug:        dd.R.DeviceMake.NameSlug,
			TokenId:         manufacturerID.Uint64(),
		},
		Verified:     dd.Verified,
		Transactions: dd.TRXHashHex,
	}

	rp.DeviceStyles = []*grpc.DeviceStyle{}
	for _, ds := range dd.R.DeviceStyles {
		rp.DeviceStyles = append(rp.DeviceStyles, &grpc.DeviceStyle{
			DeviceDefinitionId: tbl.KSUID,
			ExternalStyleId:    ds.ExternalStyleID,
			Id:                 ds.ID,
			Name:               ds.Name,
			Source:             ds.Source,
			SubModel:           ds.SubModel,
		})
	}

	rp.DeviceAttributes = []*grpc.DeviceTypeAttribute{}
	for _, da := range common.GetDeviceAttributesTyped(dd.Metadata, dd.R.DeviceType.Metadatakey) {
		rp.DeviceAttributes = append(rp.DeviceAttributes, &grpc.DeviceTypeAttribute{
			Name:        da.Name,
			Label:       da.Label,
			Description: da.Description,
			Value:       da.Value,
			Required:    da.Required,
			Type:        da.Type,
			Options:     da.Option,
		})
	}

	return rp
}
