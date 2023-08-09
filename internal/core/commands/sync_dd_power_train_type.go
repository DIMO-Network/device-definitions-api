package commands

import (
	"context"
	"database/sql"
	"errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"os"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gopkg.in/yaml.v3"
)

type SyncPowerTrainTypeCommand struct {
}

type SyncPowerTrainTypeCommandResult struct {
	Status bool
}

func (*SyncPowerTrainTypeCommand) Key() string { return "SyncPowerTrainTypeCommand" }

type SyncPowerTrainTypeCommandHandler struct {
	DBS    func() *db.ReaderWriter
	logger zerolog.Logger
}

func NewSyncPowerTrainTypeCommandHandler(dbs func() *db.ReaderWriter, logger zerolog.Logger) SyncPowerTrainTypeCommandHandler {
	return SyncPowerTrainTypeCommandHandler{DBS: dbs, logger: logger}
}

func (ch SyncPowerTrainTypeCommandHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {

	const powerTrainType = "powertrain_type"

	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		models.DeviceDefinitionWhere.DeviceTypeID.EQ(null.StringFrom("vehicle")),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}

	ch.logger.Info().Msgf("powertrain setting - found %d device definitions verified, starting process...", len(all))
	if len(all) == 0 {
		return SyncPowerTrainTypeCommandResult{false}, nil
	}

	content, err := os.ReadFile("powertrain_type_rule.yaml")
	if err != nil {
		return nil, err
	}

	var powerTrainTypeData coremodels.PowerTrainTypeRuleData
	err = yaml.Unmarshal(content, &powerTrainTypeData)
	if err != nil {
		return nil, err
	}

	for _, definition := range all {
		ch.logger.Info().Msgf("%s - Make:%s Model: %s Year: %d", definition.ID, definition.R.DeviceMake.NameSlug, definition.ModelSlug, definition.Year)

		if definition.R.DeviceType == nil {
			ch.logger.Error().Msgf("ID: %s with DeviceType is empty", definition.ID)
			continue
		}

		metadataKey := definition.R.DeviceType.Metadatakey
		var metadataAttributes map[string]any

		if err := definition.Metadata.Unmarshal(&metadataAttributes); err == nil {
			metadataAttributes = make(map[string]interface{})
			metaData := make(map[string]interface{})

			dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(definition.DeviceTypeID.String)).
				One(ctx, ch.DBS().Reader)

			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, err
				}
			}

			var deviceTypeAttributes map[string][]coremodels.GetDeviceTypeAttributeQueryResult
			if err := dt.Properties.Unmarshal(&deviceTypeAttributes); err == nil {
				for _, deviceAttribute := range deviceTypeAttributes["properties"] {
					metaData[deviceAttribute.Name] = deviceAttribute.DefaultValue
				}
			}

			metadataAttributes[metadataKey] = metaData
		}

		for key, _ := range metadataAttributes[metadataKey].(map[string]any) {

			if key == powerTrainType {
				powerTrainTypeValue := ch.resolvePowerTrainTypeByMake(ctx, powerTrainTypeData, definition)
				metadataAttributes[metadataKey].(map[string]interface{})[powerTrainType] = powerTrainTypeValue
			}
		}

		err = definition.Metadata.Marshal(metadataAttributes)
		if err != nil {
			return nil, err
		}

		if err = definition.Upsert(ctx, ch.DBS().Writer, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
			return nil, err
		}
	}

	return SyncPowerTrainTypeCommandResult{true}, nil
}

func (ch SyncPowerTrainTypeCommandHandler) resolvePowerTrainTypeByMake(ctx context.Context, powerTrainTypeData coremodels.PowerTrainTypeRuleData,
	definition *models.DeviceDefinition) string {

	for _, ptType := range powerTrainTypeData.PowerTrainTypeList {
		for _, mk := range ptType.Makes {
			if mk == definition.R.DeviceMake.NameSlug {
				if len(ptType.Models) == 0 {
					return ptType.Type
				}

				for _, model := range ptType.Models {
					if model == definition.ModelSlug {
						return ptType.Type
					}
				}

			}
		}
	}

	// Default
	defaultPowerTrainType := ""
	for _, ptType := range powerTrainTypeData.PowerTrainTypeList {
		if ptType.Default {
			defaultPowerTrainType = ptType.Type
			break
		}
	}

	vins, err := models.VinNumbers(models.VinNumberWhere.DeviceDefinitionID.EQ(definition.ID)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return defaultPowerTrainType
	}

	if len(vins) == 0 {
		return defaultPowerTrainType
	}

	vin := vins[0]

	// Resolve Drivly Data
	ch.logger.Info().Msg("Looking up PowerTrain from Drivly Data")
	if vin.DrivlyData.Valid && len(powerTrainTypeData.VincarioList) > 0 {
		var drivlyData coremodels.DrivlyData
		err = vin.DrivlyData.Unmarshal(&drivlyData)
		if err != nil {
			ch.logger.Error().Err(err)
		}

		for _, item := range powerTrainTypeData.DrivlyList {
			if len(item.Values) > 0 {
				for _, value := range item.Values {
					if value == drivlyData.FuelType {
						return item.Type
					}
				}
			}
		}

	}

	// Resolve Vincario Data
	ch.logger.Info().Msg("Looking up PowerTrain from Vincario Data")
	if vin.VincarioData.Valid && len(powerTrainTypeData.VincarioList) > 0 {
		var vincarioData coremodels.VincarioData
		err = vin.DrivlyData.Unmarshal(&vincarioData)
		if err != nil {
			ch.logger.Error().Err(err)
		}

		for _, item := range powerTrainTypeData.VincarioList {
			if len(item.Values) > 0 {
				for _, value := range item.Values {
					if value == vincarioData.FuelType {
						return item.Type
					}
				}
			}
		}
	}

	return defaultPowerTrainType
}
