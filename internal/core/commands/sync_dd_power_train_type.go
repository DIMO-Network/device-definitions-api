package commands

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SyncPowerTrainTypeCommand struct {
	ForceUpdate  bool
	DeviceTypeID string
}

type SyncPowerTrainTypeCommandResult struct {
	Status bool
}

func (*SyncPowerTrainTypeCommand) Key() string { return "SyncPowerTrainTypeCommand" }

type SyncPowerTrainTypeCommandHandler struct {
	DBS                   func() *db.ReaderWriter
	logger                zerolog.Logger
	powerTrainTypeService services.PowerTrainTypeService
}

func NewSyncPowerTrainTypeCommandHandler(dbs func() *db.ReaderWriter, logger zerolog.Logger, powerTrainTypeService services.PowerTrainTypeService) SyncPowerTrainTypeCommandHandler {
	return SyncPowerTrainTypeCommandHandler{DBS: dbs, logger: logger, powerTrainTypeService: powerTrainTypeService}
}

const vehicleTypeMDKey = "vehicle_info" // expected key for correct device type

func (ch SyncPowerTrainTypeCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	command := query.(*SyncPowerTrainTypeCommand)

	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		models.DeviceDefinitionWhere.DeviceTypeID.EQ(null.StringFrom(command.DeviceTypeID)),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}

	ch.logger.Info().Msgf("powertrain setting Force (%t) - found %d device definitions verified, starting process...", command.ForceUpdate, len(all))
	if len(all) == 0 {
		return SyncPowerTrainTypeCommandResult{false}, nil
	}

	for _, definition := range all {
		//ch.logger.Info().Msgf("%s - Make:%s Model: %s Year: %d", definition.ID, definition.R.DeviceMake.NameSlug, definition.ModelSlug, definition.Year)
		needsUpdate := false

		if definition.R.DeviceType == nil {
			ch.logger.Error().Msgf("ID: %s with DeviceType is empty", definition.ID)
			continue
		}

		metadataKey := definition.R.DeviceType.Metadatakey
		if metadataKey != vehicleTypeMDKey {
			ch.logger.Info().Msgf("skipping dd id: %s because not vehicle device type: %s", definition.ID, metadataKey)
			continue
		}
		var metadataAttributes map[string]any

		if err = definition.Metadata.Unmarshal(&metadataAttributes); err == nil {
			metaData := make(map[string]interface{})
			if metadataAttributes == nil {
				metadataAttributes = make(map[string]interface{})
				var deviceTypeAttributes map[string][]coremodels.GetDeviceTypeAttributeQueryResult
				if err := definition.R.DeviceType.Properties.Unmarshal(&deviceTypeAttributes); err == nil {
					for _, deviceAttribute := range deviceTypeAttributes["properties"] {
						metaData[deviceAttribute.Name] = deviceAttribute.DefaultValue
					}
				}

				metadataAttributes[metadataKey] = metaData
			}

			// Validate format
			if _, ok := metadataAttributes[metadataKey]; ok {
				var powerTrainTypeValue *string
				hasPowerTrainType := false
				for key, value := range metadataAttributes[metadataKey].(map[string]interface{}) {
					if key == common.PowerTrainType {
						hasPowerTrainType = true
						if strValue, isString := value.(string); isString {
							powerTrainTypeValue = &strValue
						}
						if powerTrainTypeValue == nil || *powerTrainTypeValue == "" || *powerTrainTypeValue == "null" {
							powerTrainTypeValue, err = ch.powerTrainTypeService.ResolvePowerTrainType(ctx, definition.R.DeviceMake.NameSlug, definition.ModelSlug, definition)
							if err != nil {
								ch.logger.Error().Err(err).Stack().Msg("failed to ResolvePowerTrainType")
								return nil, err
							}
							needsUpdate = true
						} else {
							if command.ForceUpdate {
								powerTrainTypeValue, err = ch.powerTrainTypeService.ResolvePowerTrainType(ctx, definition.R.DeviceMake.NameSlug, definition.ModelSlug, definition)
								if err != nil {
									ch.logger.Error().Err(err).Stack().Msg("failed to ResolvePowerTrainType")
									return nil, err
								}
								ch.logger.Info().Msgf("Current Powertraintype:%s | New Powertraintype: %s", metadataAttributes[metadataKey].(map[string]interface{})[common.PowerTrainType], *powerTrainTypeValue)
								needsUpdate = true
							}
						}
						break
					}
				}

				if !hasPowerTrainType {
					powerTrainTypeValue, err = ch.powerTrainTypeService.ResolvePowerTrainType(ctx, definition.R.DeviceMake.NameSlug, definition.ModelSlug, definition)

					if err != nil {
						ch.logger.Error().Err(err).Stack().Msg("failed to ResolvePowerTrainType")
						return nil, err
					}
					needsUpdate = true
				}
				metadataAttributes[metadataKey].(map[string]interface{})[common.PowerTrainType] = powerTrainTypeValue
			}
		}

		err = definition.Metadata.Marshal(metadataAttributes)
		if err != nil {
			return nil, err
		}
		if needsUpdate {
			if _, err = definition.Update(ctx, ch.DBS().Writer, boil.Whitelist(models.DeviceDefinitionColumns.UpdatedAt, models.DeviceDefinitionColumns.Metadata)); err != nil {
				return nil, err
			}
		}

	}

	return SyncPowerTrainTypeCommandResult{true}, nil
}
