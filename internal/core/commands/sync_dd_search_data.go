package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/elastic"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SyncSearchDataCommand struct {
}

type SyncSearchDataCommandResult struct {
	Status bool
}

func (*SyncSearchDataCommand) Key() string { return "SyncSearchDataCommand" }

type SyncSearchDataCommandHandler struct {
	DBS    func() *db.ReaderWriter
	esSvc  elastic.SearchService
	logger zerolog.Logger
}

func NewSyncSearchDataCommandHandler(dbs func() *db.ReaderWriter, esSvc elastic.SearchService, logger zerolog.Logger) SyncSearchDataCommandHandler {
	return SyncSearchDataCommandHandler{DBS: dbs, esSvc: esSvc, logger: logger}
}

func (ch SyncSearchDataCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	existingEngines, err := ch.esSvc.GetEngines()
	if err != nil {
		return nil, err
	}
	ch.logger.Info().Msgf("found existing engines: %d", len(existingEngines.Results))

	// get all devices from DB.
	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}

	ch.logger.Info().Msgf("found %d device definitions verified, starting process...", len(all))

	if len(all) == 0 {
		return nil, errors.New("0 items found to index, stopping")
	}
	metaEngineName := ch.esSvc.GetMetaEngineName()
	docs := make([]elastic.DeviceDefinitionSearchDoc, len(all))
	for i, definition := range all {
		sd := fmt.Sprintf("%d %s %s", definition.Year, definition.R.DeviceMake.Name, definition.Model)
		sm := common.SubModelsFromStylesDB(definition.R.DeviceStyles)
		for i2, s := range sm {
			sm[i2] = sd + " " + s
		}
		docs[i] = elastic.DeviceDefinitionSearchDoc{
			ID:            definition.ID,
			SearchDisplay: sd,
			Make:          definition.R.DeviceMake.Name,
			Model:         definition.Model,
			Year:          int(definition.Year),
			SubModels:     sm,
			ImageURL:      definition.ImageURL.String,
			MakeSlug:      definition.R.DeviceMake.NameSlug,
			ModelSlug:     definition.ModelSlug,
		}
	}
	ch.logger.Info().Msgf("completed building list of docs to index, count: %d", len(docs))

	tempEngineName := fmt.Sprintf("%s-%s", metaEngineName, time.Now().Format("2006-01-02t15-04"))
	tempEngine, err := ch.esSvc.CreateEngine(tempEngineName, nil)
	if err != nil {
		return nil, err
	}
	ch.logger.Info().Msgf("created engine %s", tempEngine.Name)
	err = ch.esSvc.CreateDocumentsBatched(docs, tempEngine.Name)
	if err != nil {
		return nil, err
	}
	ch.logger.Info().Msgf("created documents in engine %s", tempEngine.Name)

	var metaEngine *elastic.EngineDetail
	var previousTempEngines []string
	// look for existing meta engine, and any previous core engines that should be removed.
	for _, result := range existingEngines.Results {
		if result.Name == metaEngineName && *result.Type == "meta" {
			metaEngine = &result
			ch.logger.Info().Msgf("found existing meta engine: %+v", *metaEngine)
		}
		if strings.Contains(result.Name, metaEngineName+"-") && *result.Type == "default" {
			previousTempEngines = append(previousTempEngines, result.Name)
			ch.logger.Info().Msgf("found previous device defs engine: %s. It will be removed", result.Name)
		}
	}
	if metaEngine == nil {
		_, err = ch.esSvc.CreateEngine(metaEngineName, &tempEngineName)
		if err != nil {
			return nil, err
		}
		ch.logger.Info().Msg("created meta engine with temp engine assigned.")
	} else {
		_, err = ch.esSvc.AddSourceEngineToMetaEngine(tempEngineName, metaEngineName)
		if err != nil {
			return nil, err
		}
		ch.logger.Info().Msgf("added source %s to meta engine %s", tempEngine.Name, metaEngineName)
		for _, prev := range previousTempEngines {
			// loop over all previous ones
			if common.Contains(metaEngine.SourceEngines, prev) {
				_, err = ch.esSvc.RemoveSourceEngine(prev, metaEngineName)
				if err != nil {
					return nil, err
				}
				ch.logger.Info().Msgf("removed previous source engine %s from %s", prev, metaEngineName)
			}

			err = ch.esSvc.DeleteEngine(prev)
			if err != nil {
				return nil, err
			}
			ch.logger.Info().Msgf("delete engine: %s", prev)
		}
	}
	err = ch.esSvc.UpdateSearchSettingsForDeviceDefs(tempEngineName)
	if err != nil {
		return nil, err
	}
	err = ch.esSvc.UpdateSearchSettingsForDeviceDefs(metaEngineName)
	if err != nil {
		return nil, err
	}
	ch.logger.Info().Msg("completed loading device definitions into search index ok")

	return SyncSearchDataCommandResult{true}, nil
}
