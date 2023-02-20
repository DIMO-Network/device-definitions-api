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
	vinRepository      repositories.VINRepository
}

func NewSyncVinNumbersCommand(dbs func() *db.ReaderWriter,
	vinDecodingService services.VINDecodingService,
	repository repositories.DeviceDefinitionRepository,
	vinRepository repositories.VINRepository,
	logger *zerolog.Logger) SyncVinNumbersCommandHandler {
	return SyncVinNumbersCommandHandler{
		dbs:                dbs,
		vinDecodingService: vinDecodingService,
		logger:             logger,
		repository:         repository,
		vinRepository:      vinRepository,
	}
}

func (dc SyncVinNumbersCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*SyncVinNumbersCommand)
	for _, vinNumber := range command.VINNumbers {
		if len(vinNumber) != 17 {
			dc.logger.Warn().Str("vin", vinNumber).Msgf("invalid vin %s", vinNumber)
			continue
		}

		vin := shared.VIN(vinNumber)
		year := int32(vin.Year()) // needs to be updated for newer years
		wmi := vin.Wmi()

		localLog := dc.logger.With().
			Str("vin", vin.String()).
			Str("handler", query.Key()).
			Str("year", fmt.Sprintf("%d", year)).
			Logger()

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

			dbWMI, err := dc.vinRepository.GetOrCreateWMI(ctx, wmi, vinInfo.Make)
			if err != nil {
				dc.logger.Error().Str("vin", vin.String()).Msgf("invalid vin %s", vin.String())
				continue
			}

			// now match the model for the dd id
			dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.DeviceMakeID.EQ(dbWMI.DeviceMakeID),
				models.DeviceDefinitionWhere.Year.EQ(int16(year)),
				models.DeviceDefinitionWhere.ModelSlug.EQ(common.SlugString(vinInfo.Model))).
				One(ctx, dc.dbs().Reader)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					dd, err = dc.repository.GetOrCreate(ctx,
						vinInfo.Source,
						common.SlugString(vinInfo.Model+vinInfo.Year),
						dbWMI.DeviceMakeID,
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
						Source:             vinInfo.Source,
						SubModel:           vinInfo.SubModel,
					}
					_ = style.Insert(ctx, dc.dbs().Writer, boil.Infer())
					localLog.Info().
						Msgf("creating new device_style as did not find one for: %s", common.SlugString(vinInfo.StyleName))
				}
			}

			vinDecodeNumber = &models.VinNumber{
				Vin:                vin.String(),
				DeviceDefinitionID: dd.ID,
				DeviceMakeID:       dd.DeviceMakeID,
				Wmi:                wmi,
				VDS:                vin.VDS(),
				Vis:                vin.VIS(),
				CheckDigit:         vin.CheckDigit(),
				SerialNumber:       vin.SerialNumber(),
			}
			if err = vinDecodeNumber.Insert(ctx, dc.dbs().Writer, boil.Infer()); err != nil {
				localLog.Err(err).
					Str("vin", vin.String()).
					Str("device_definition_id", dd.ID).
					Str("device_make_id", dd.DeviceMakeID).
					Msgf("failed to insert vin_numbers: %s", vin.String())
			}
		}

	}

	return &SyncVinNumbersCommandResult{true}, nil
}
