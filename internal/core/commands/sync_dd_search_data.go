package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SyncSearchDataCommand struct {
}

type SyncSearchDataCommandResult struct {
}

func (*SyncSearchDataCommand) Key() string { return "SyncSearchDataCommand" }

type SyncSearchDataCommandHandler struct {
	DBS   func() *db.ReaderWriter
	esSvc gateways.ElasticSearchService
}

func NewSyncSearchDataCommandHandler(dbs func() *db.ReaderWriter, esSvc gateways.ElasticSearchService) SyncSearchDataCommandHandler {
	return SyncSearchDataCommandHandler{DBS: dbs, esSvc: esSvc}
}

func (ch SyncSearchDataCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	//metaEngineName := ""
	existingEngines, err := ch.esSvc.GetEngines()
	if err != nil {
		return nil, err
	}
	fmt.Printf("found existing engines: %d", len(existingEngines.Results))

	// get all devices from DB.
	all, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}

	fmt.Printf("found %d device definitions verified", len(all))
	if len(all) == 0 {
		return nil, errors.New("0 items found to index, stopping")
	}

	docs := make([]gateways.DeviceDefinitionSearchDoc, len(all))
	for i, definition := range all {
		sd := fmt.Sprintf("%d %s %s", definition.Year, definition.R.DeviceMake.Name, definition.Model)
		sm := common.SubModelsFromStylesDB(definition.R.DeviceStyles)
		for i2, s := range sm {
			sm[i2] = sd + " " + s
		}
		docs[i] = gateways.DeviceDefinitionSearchDoc{
			ID:            definition.ID,
			SearchDisplay: sd,
			Make:          definition.R.DeviceMake.Name,
			Model:         definition.Model,
			Year:          int(definition.Year),
			SubModels:     sm,
			ImageURL:      definition.ImageURL.String,
		}
	}

	tempEngineName := fmt.Sprintf("%s-%s", ch.esSvc.GetMetaEngineName(), time.Now().Format("2006-01-02t15-04"))
	tempEngine, err := ch.esSvc.CreateEngine(tempEngineName, nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("created engine %s", tempEngine.Name)
	err = ch.esSvc.CreateDocumentsBatched(docs, tempEngine.Name)
	if err != nil {
		return nil, err
	}
	fmt.Printf("created documents in engine %s", tempEngine.Name)

	var metaEngine *gateways.EngineDetail
	var previousTempEngines []string
	// look for existing meta engine, and any previous core engines that should be removed.
	for _, result := range existingEngines.Results {
		if result.Name == ch.esSvc.GetMetaEngineName() && *result.Type == "meta" {
			metaEngine = &result
			fmt.Printf("found existing meta engine: %+v", *metaEngine)
		}
		if strings.Contains(result.Name, ch.esSvc.GetMetaEngineName()+"-") && *result.Type == "default" {
			previousTempEngines = append(previousTempEngines, result.Name)
			fmt.Printf("found previous device defs engine: %s. It will be removed", result.Name)
		}
	}
	if metaEngine == nil {
		_, err = ch.esSvc.CreateEngine(ch.esSvc.GetMetaEngineName(), &tempEngineName)
		if err != nil {
			return nil, err
		}
		fmt.Printf("created meta engine with temp engine assigned.")
	} else {
		_, err = ch.esSvc.AddSourceEngineToMetaEngine(tempEngineName, ch.esSvc.GetMetaEngineName())
		if err != nil {
			return nil, err
		}
		fmt.Printf("added source %s to meta engine %s", tempEngine.Name, ch.esSvc.GetMetaEngineName())
		for _, prev := range previousTempEngines {
			// loop over all previous ones
			if common.Contains(metaEngine.SourceEngines, prev) {
				_, err = ch.esSvc.RemoveSourceEngine(prev, ch.esSvc.GetMetaEngineName())
				if err != nil {
					return nil, err
				}
				fmt.Printf("removed previous source engine %s from %s", prev, ch.esSvc.GetMetaEngineName())
			}

			err = ch.esSvc.DeleteEngine(prev)
			if err != nil {
				return nil, err
			}
			fmt.Printf("delete engine: %s", prev)
		}
	}
	err = ch.esSvc.UpdateSearchSettingsForDeviceDefs(tempEngineName)
	if err != nil {
		return nil, err
	}
	err = ch.esSvc.UpdateSearchSettingsForDeviceDefs(ch.esSvc.GetMetaEngineName())
	if err != nil {
		return nil, err
	}
	fmt.Printf("completed load ok")

	return SyncSearchDataCommandResult{}, nil
}
