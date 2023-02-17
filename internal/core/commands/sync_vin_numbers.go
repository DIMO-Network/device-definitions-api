package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type SyncVinNumbersCommand struct {
	VINNumbers []string
}

type SyncVinNumbersCommandResult struct {
	Status bool
}

func (*SyncVinNumbersCommand) Key() string { return "SyncVinNumbersCommand" }

type SyncVinNumbersCommandHandler struct {
	dbs                func() *db.ReaderWriter
	vinDecodingService services.VINDecodingService
	logger             *zerolog.Logger
	repository         repositories.DeviceDefinitionRepository
}

func NewSyncVinNumbersCommand(dbs func() *db.ReaderWriter,
	vinDecodingService services.VINDecodingService,
	repository repositories.DeviceDefinitionRepository,
	logger *zerolog.Logger) SyncVinNumbersCommandHandler {
	return SyncVinNumbersCommandHandler{
		dbs:                dbs,
		vinDecodingService: vinDecodingService,
		logger:             logger,
		repository:         repository,
	}
}

func (dc SyncVinNumbersCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*SyncVinNumbersCommand)
	for _, vinNumber := range command.VINNumbers {
		if len(vinNumber) != 17 {
			dc.logger.Warn().Str("vin", vinNumber).Msgf("invalid vin %s", vinNumber)
			continue
		}

		var deviceMakeID = ""
		vin := shared.VIN(vinNumber)
		year := int32(vin.Year()) // needs to be updated for newer years
		wmi := vin.Wmi()

		localLog := dc.logger.With().
			Str("vin", vin.String()).
			Str("handler", query.Key()).
			Str("year", fmt.Sprintf("%d", year)).
			Logger()

		dbWMI, err := models.FindWmi(ctx, dc.dbs().Reader, wmi)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			dc.logger.Error().Str("vin", vin.String()).Msgf("invalid vin %s", vin.String())
			continue
		}
		if dbWMI != nil {
			deviceMakeID = dbWMI.DeviceMakeID
		}
		vinDecodeNumber, err := models.FindVinNumber(ctx, dc.dbs().Reader, vin.String())
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				localLog.Debug().Msg("No rows to decode vin numbers from db")
			}
		}

		if vinDecodeNumber == nil {
			dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).
				One(ctx, dc.dbs().Reader)
			if err != nil {
				localLog.Err(err)
				continue
			}

			vinInfo, err := dc.vinDecodingService.GetVIN(vin.String(), dt)
			if err != nil {
				localLog.Err(err).Msg("failed to decode vin from drivly")
				continue
			}

			// get the make from the vinInfo if no WMI found
			if dbWMI == nil {
				deviceMake, err := models.
					DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(common.SlugString(vinInfo.Make))).
					One(ctx, dc.dbs().Reader)

				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						localLog.Warn().Msgf("failed to find make from vin decode with name slug: %s", common.SlugString(vinInfo.Make))
					} else {
						localLog.Err(err)
						continue
					}
				}

				if deviceMake != nil {
					deviceMakeID = deviceMake.ID
					// insert the WMI
					dbWMI = &models.Wmi{
						Wmi:          wmi,
						DeviceMakeID: deviceMake.ID,
					}
					if err = dbWMI.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
						localLog.Err(err).Str("deviceMakeId", deviceMake.ID).Msgf("failed to insert wmi: %s", wmi)
					}
				}
			}

			// now match the model for the dd id
			dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(deviceMakeID),
				models.DeviceDefinitionWhere.Year.EQ(int16(year)),
				models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
				One(ctx, dc.dbs().Reader)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					dd, err = dc.repository.GetOrCreate(ctx,
						vinInfo.Source,
						common.SlugString(vinInfo.Model+vinInfo.Year),
						deviceMakeID,
						vinInfo.Model,
						int(year),
						common.DefaultDeviceType,
						vinInfo.MetaData,
						true,
						"")
					if err != nil {
						localLog.Err(err)
						continue
					}
					localLog.Info().Msgf("creating new DD as did not find DD from vin decode with model slug: %s", common.SlugString(vinInfo.Model))
				} else {
					return nil, err
				}
			}
			if dd != nil {
				// match style
				var style, err = models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID),
					models.DeviceStyleWhere.SubModel.EQ(vinInfo.SubModel),
					models.DeviceStyleWhere.Name.EQ(vinInfo.StyleName)).One(ctx, dc.dbs().Reader)
				if errors.Is(err, sql.ErrNoRows) {
					// insert
					style = &models.DeviceStyle{
						ID:                 ksuid.New().String(),
						DeviceDefinitionID: dd.ID,
						Name:               vinInfo.StyleName,
						ExternalStyleID:    common.SlugString(vinInfo.StyleName),
						Source:             "drivly",
						SubModel:           vinInfo.SubModel,
					}
					_ = style.Insert(ctx, dc.dbs().Writer, boil.Infer())
					localLog.Info().
						Msgf("creating new device_style as did not find one for: %s", common.SlugString(vinInfo.StyleName))
				}
			}

		}

	}

	return &SyncVinNumbersCommandResult{true}, nil
}
