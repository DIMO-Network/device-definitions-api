package queries

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
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

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter, drivlyAPISvc gateways.DrivlyAPIService, logger *zerolog.Logger) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		DBS:          dbs,
		drivlyApiSvc: drivlyAPISvc,
		logger:       logger,
	}
}

// todo write a test for this once have DB structure
// todo add grpc decode in the api folder to wire this up.

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*DecodeVINQuery)
	if len(qry.VIN) != 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
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
			// refactor with repo? service / command?
			dd = &models.DeviceDefinition{
				ID:           ksuid.New().String(),
				Model:        vinInfo.Model,
				Year:         int16(resp.Year),
				Metadata:     null.JSON{}, // todo build attributes
				Verified:     true,
				DeviceMakeID: resp.DeviceMakeId,
				ModelSlug:    common.SlugString(vinInfo.Model),
				DeviceTypeID: null.StringFrom("vehicle"),
				ExternalIds:  null.JSON{},
			}
			err = dd.Insert(ctx, dc.DBS().Writer, boil.Infer())
			if err != nil {
				return nil, err
			}
			localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", common.SlugString(vinInfo.Model))
		} else {
			return nil, err
		}
	}
	if dd != nil {
		resp.DeviceDefinitionId = dd.ID
		// match style
		style, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
			models.DeviceStyleWhere.SubModel.EQ(vinInfo.SubModel),
			models.DeviceStyleWhere.Name.EQ(buildStyleName(vinInfo))).One(ctx, dc.DBS().Reader)
		if err == nil {
			resp.DeviceStyleId = style.ID
		} else if errors.Is(err, sql.ErrNoRows) {
			// insert
			style = &models.DeviceStyle{
				ID:                 ksuid.New().String(),
				DeviceDefinitionID: dd.ID,
				Name:               buildStyleName(vinInfo),
				ExternalStyleID:    common.SlugString(buildStyleName(vinInfo)),
				Source:             "drivly",
				SubModel:           vinInfo.SubModel,
			}
			_ = style.Insert(ctx, dc.DBS().Writer, boil.Infer())
		}
		// todo update the metadata if different? add powertrain - but this can be style specific
	}

	return resp, nil
}

func buildStyleName(vinInfo *gateways.VINInfoResponse) string {
	return vinInfo.Trim + " " + vinInfo.SubModel
}
