package queries

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DecodeVINQueryHandler struct {
	DBS          func() *db.ReaderWriter
	drivlyApiSvc gateways.DrivlyAPIService
	logger       *zerolog.Logger
}

type DecodeVINQuery struct {
	VIN string `json:"vin"`
}

func (*DecodeVINQuery) Key() string { return "DecodeVINQuery" }

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter, settings *config.Settings, logger *zerolog.Logger) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		DBS:          dbs,
		drivlyApiSvc: gateways.NewDrivlyAPIService(settings),
		logger:       logger,
	}
}

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*DecodeVINQuery)
	if len(qry.VIN) != 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
	// todo write a test for this once have DB structure
	resp := &p_grpc.DecodeVINResponse{}
	// get the year
	vin := shared.VIN(qry.VIN)
	resp.Year = int32(vin.Year()) // needs to be updated for newer years
	// todo: we could decode tesla on our own
	// get the make
	wmi := qry.VIN[0:3]
	dbWMI, err := models.FindWmi(ctx, dc.DBS().Reader, wmi)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if dbWMI != nil {
		resp.DeviceMakeId = dbWMI.DeviceMakeID
	}

	localLog := dc.logger.With().Str("vin", qry.VIN).Str("handler", qry.Key()).Str("year", string(resp.Year)).Logger()
	// not yet - lookup the device definition by rest of info - look at our existing vins
	// for now always call drivly to decode
	vinInfo, err := dc.drivlyApiSvc.GetVINInfo(vin.String())
	// get the make from the vinInfo if no WMI found
	if dbWMI == nil {
		deviceMake, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(common.SlugString(vinInfo.Make))).One(ctx, dc.DBS().Reader)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				localLog.Warn().Msgf("failed to find make from vin decode with name slug: %s", common.SlugString(vinInfo.Make))
			} else {
				return nil, err
			}
		}
		if deviceMake != nil {
			resp.DeviceMakeId = deviceMake.ID
			// insert the WMI
			dbWMI = &models.Wmi{
				Wmi:          wmi,
				DeviceMakeID: deviceMake.ID,
			}
			if err = dbWMI.Insert(ctx, dc.DBS().Writer, boil.Infer()); err != nil {
				localLog.Err(err).Str("deviceMakeId", deviceMake.ID).Msgf("failed to insert wmi: %s", wmi)
			}
		}
	}
	// now match the model for the dd id
	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(resp.DeviceMakeId),
		models.DeviceDefinitionWhere.Year.EQ(int16(resp.Year)),
		models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
		One(ctx, dc.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			localLog.Warn().Msgf("failed to find device_definition from vin decode with model slug: %s", common.SlugString(vinInfo.Model))
		} else {
			return nil, err
		}
	}
	if dd != nil {
		resp.DeviceDefinitionId = dd.ID
	}

	return resp, nil
}
