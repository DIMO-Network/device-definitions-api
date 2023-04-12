package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SyncSmartCartForwardCompatibilityCommand struct {
}

type SyncSmartCartForwardCompatibilityCommandResult struct {
	Status bool
}

func (*SyncSmartCartForwardCompatibilityCommand) Key() string {
	return "SyncSmartCartForwardCompatibilityCommand"
}

type SyncSmartCartForwardCompatibilityCommandHandler struct {
	DBS        func() *db.ReaderWriter
	scSvc      gateways.SmartCarService
	Repository repositories.DeviceDefinitionRepository
}

func NewSyncSmartCartForwardCompatibilityCommandHandler(dbs func() *db.ReaderWriter, scSvc gateways.SmartCarService, repository repositories.DeviceDefinitionRepository) SyncSmartCartForwardCompatibilityCommandHandler {
	return SyncSmartCartForwardCompatibilityCommandHandler{DBS: dbs, scSvc: scSvc, Repository: repository}
}

func (ch SyncSmartCartForwardCompatibilityCommandHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {
	integrationID, err := ch.scSvc.GetOrCreateSmartCarIntegration(ctx)

	if err != nil {
		return nil, err
	}

	deviceDefs, err := models.DeviceDefinitions(
		qm.InnerJoin("device_definitions_api.device_integrations di on di.device_definition_id = device_definitions.id"),
		qm.Where("di.integration_id = ?", integrationID),
		qm.Load(models.DeviceDefinitionRels.DeviceMake)).
		All(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, err
	}
	fmt.Printf("found %d device definitions with smartcar integration\n", len(deviceDefs))

	lastMM := ""
	lastYear := int16(0)
	// meant to be used at the end of each loop to update the "last" values
	funcLastValues := func(dd *models.DeviceDefinition) {
		lastYear = dd.Year
		lastMM = dd.R.DeviceMake.Name + dd.Model
	}
	// year will be descending
	for _, dd := range deviceDefs {
		thisMM := dd.R.DeviceMake.Name + dd.Model
		if lastMM == thisMM {
			// we care about year gaps
			yearDiff := lastYear - dd.Year
			if yearDiff > 1 {
				// we have a gap
				fmt.Printf("%s found a year gap of %d...\n", thisMM, yearDiff)
				for i := int16(1); i < yearDiff; i++ {
					gapDd, err := ch.Repository.GetByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year+i), true)
					if errors.Is(err, sql.ErrNoRows) {
						continue // this continues internal loop, so funcLastValues will still get set at end of outer loop
					}
					if err != nil {
						return nil, err
					}
					// found a record that needs to be attached to integration
					if len(gapDd.R.DeviceIntegrations) == 0 {
						fmt.Printf("found device def for year gap %s, inserting smartcar device_integration\n", common.PrintMMY(gapDd, common.Green, true))
						diGap := models.DeviceIntegration{
							DeviceDefinitionID: gapDd.ID,
							IntegrationID:      integrationID,
							Region:             common.AmericasRegion.String(), // default
						}
						err = diGap.Insert(ctx, ch.DBS().Writer, boil.Infer())
						if err != nil {
							return nil, errors.Wrap(err, "error inserting device_integration")
						}
					} else {
						fmt.Printf("%s already had an integration set\n", common.PrintMMY(gapDd, common.Red, true))
					}
				}
			}
		} else {
			// this should mean we are back at the start of a new make/model starting at highest year
			nextYearDd, err := ch.Repository.GetByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year+1), true)
			if errors.Is(err, sql.ErrNoRows) {
				funcLastValues(dd)
				continue
			}
			if err != nil {
				return nil, err
			}
			// does it have any integrations?
			if len(nextYearDd.R.DeviceIntegrations) == 0 {
				// attach smartcar integration
				fmt.Printf("found device def for future year %s, that does not have any integrations. inserting device_integration\n", common.PrintMMY(nextYearDd, common.Green, true))
				diGap := models.DeviceIntegration{
					DeviceDefinitionID: nextYearDd.ID,
					IntegrationID:      integrationID,
					Region:             common.AmericasRegion.String(), // default
				}
				err = diGap.Insert(ctx, ch.DBS().Writer, boil.Infer())
				if err != nil {
					return nil, errors.Wrap(err, "error inserting device_integration")
				}
			}
		}

		funcLastValues(dd)
	}

	return SyncSmartCartForwardCompatibilityCommandResult{Status: true}, nil
}
