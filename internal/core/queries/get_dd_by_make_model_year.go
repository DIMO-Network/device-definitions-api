package queries

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/pkg/db"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type GetDeviceDefinitionByMakeModelYearQuery struct {
	Make  string `json:"make" validate:"required"`
	Model string `json:"model" validate:"required"`
	Year  int    `json:"year" validate:"required"`
}

func (*GetDeviceDefinitionByMakeModelYearQuery) Key() string {
	return "GetDeviceDefinitionByMakeModelYearQuery"
}

type GetDeviceDefinitionByMakeModelYearQueryHandler struct {
	dbs        func() *db.ReaderWriter
	onChainSvc gateways.DeviceDefinitionOnChainService
}

func NewGetDeviceDefinitionByMakeModelYearQueryHandler(onChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter) GetDeviceDefinitionByMakeModelYearQueryHandler {
	return GetDeviceDefinitionByMakeModelYearQueryHandler{
		onChainSvc: onChainSvc,
		dbs:        dbs,
	}
}

func (ch GetDeviceDefinitionByMakeModelYearQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	qry := query.(*GetDeviceDefinitionByMakeModelYearQuery)
	makeSlug := stringutils.SlugString(qry.Make)
	manufacturer, err := ch.onChainSvc.GetManufacturer(ctx, makeSlug, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}

	manufacturerID := types.NewNullDecimal(decimal.New(int64(manufacturer.TokenID), 0))
	definitions, err := ch.onChainSvc.GetDeviceDefinitions(ctx, manufacturerID, "", qry.Model, qry.Year, 0, 100)
	if err != nil {
		return nil, err
	}

	if len(definitions) == 0 {
		return nil, &exceptions.NotFoundError{
			Err: fmt.Errorf("could not find device definition with MMY: %s %s %d", qry.Make, qry.Model, qry.Year),
		}
	}
	dm, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(makeSlug)).One(ctx, ch.dbs().Reader)
	if err != nil {
		return nil, err
	}
	dss, _ := models.DeviceStyles(models.DeviceStyleWhere.DefinitionID.EQ(definitions[0].ID)).All(ctx, ch.dbs().Reader)
	trx, _ := models.DefinitionTransactions(models.DefinitionTransactionWhere.DefinitionID.EQ(definitions[0].ID)).All(ctx, ch.dbs().Reader)

	queryResult, err := common.BuildFromDeviceDefinitionToQueryResult(&definitions[0], coremodels.ConvertDeviceMakeFromDB(dm), dss, trx)
	if err != nil {
		return nil, err
	}

	return queryResult, nil
}
