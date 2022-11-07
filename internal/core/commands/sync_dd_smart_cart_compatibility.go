package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SyncSmartCartCompatibilityCommand struct {
}

type SyncSmartCartCompatibilityCommandResult struct {
	Status bool
}

func (*SyncSmartCartCompatibilityCommand) Key() string { return "SyncSmartCartCompatibilityCommand" }

type SyncSmartCartCompatibilityCommandHandler struct {
	DBS        func() *db.ReaderWriter
	scSvc      gateways.SmartCarService
	Repository repositories.DeviceDefinitionRepository
}

func NewSyncSmartCartCompatibilityCommandHandler(dbs func() *db.ReaderWriter, scSvc gateways.SmartCarService, repository repositories.DeviceDefinitionRepository) SyncSmartCartCompatibilityCommandHandler {
	return SyncSmartCartCompatibilityCommandHandler{DBS: dbs, scSvc: scSvc, Repository: repository}
}

// Handle adds device_integrations for any makes and years that are found on the smartcar website
func (ch SyncSmartCartCompatibilityCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	AmericasRegion := "Americas"
	EuropeRegion := "Europe"

	data, err := ch.scSvc.GetSmartCarVehicleData()
	if err != nil {
		return nil, err
	}

	scIntegrationID, err := ch.scSvc.GetOrCreateSmartCarIntegration(ctx)
	if err != nil {
		return nil, err
	}

	for shortRegion, data := range data.Result.Data.AllMakesTable.Edges[0].Node.CompatibilityData {
		region := ""
		switch shortRegion {
		case "US":
			region = AmericasRegion
		case "EU":
			region = EuropeRegion
		default:
			continue
		}

		//regionLogger := logger.With().Str("region", region).Logger()
		sunsetSkipCount := 0

		for _, datum := range data {
			if datum.Name == "All makes" {
				for i, row := range datum.Rows {
					mkName := row[0].Text
					if mkName == nil {
						fmt.Printf("No make name at row %d", i)
						continue
					}

					//mkLogger := regionLogger.With().Str("make", *mkName).Logger()
					rangeStr := row[0].Subtext
					if rangeStr == nil || *rangeStr == "" {
						//mkLogger.Error().Msg("Empty year range string, skipping manufacturer")
						continue
					}
					// Currently this describes Hyundai and Nissan.
					if strings.HasSuffix(*rangeStr, " (contact us)") {
						continue
					}
					startYear, err := strconv.Atoi((*rangeStr)[:len(*rangeStr)-1])
					if err != nil {
						//mkLogger.Err(err).Msg("Couldn't parse range string, skipping")
						continue
					}
					if startYear < 2012 {
						//mkLogger.Error().Msgf("Start year %d is suspiciously low, skipping", startYear)
						continue
					}

					dbMk, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(*mkName)).One(ctx, ch.DBS().Writer)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							//mkLogger.Warn().Msg("No make with this name found in the database, skipping")
							continue
						}
						return nil, fmt.Errorf("database failure: %w", err)
					}
					dds, err := dbMk.DeviceDefinitions(
						qm.LeftOuterJoin(models.TableNames.DeviceIntegrations+" ON "+models.DeviceIntegrationTableColumns.DeviceDefinitionID+" = "+models.DeviceDefinitionTableColumns.ID+" AND "+models.DeviceIntegrationTableColumns.Region+" = ?", region),
						qm.Where(models.TableNames.DeviceIntegrations+" IS NULL"),
						models.DeviceDefinitionWhere.Year.GTE(int16(startYear)),
					).All(ctx, ch.DBS().Writer)
					if err != nil {
						return nil, fmt.Errorf("database error: %w", err)
					}

					if len(dds) == 0 {
						continue
					}
					//mkLogger.Info().Msgf("Planning to insert %d compatibility records from %d onward", len(dds), startYear)

					for _, dd := range dds {
						if dd.Year < 2017 {
							// skipping as likelihood of being a 3g sunset vehicle
							sunsetSkipCount++
							continue
						}
						if err := dd.AddDeviceIntegrations(ctx, ch.DBS().Writer.DB, true, &models.DeviceIntegration{
							DeviceDefinitionID: dd.ID,
							IntegrationID:      scIntegrationID,
							Region:             region,
						}); err != nil {
							return nil, fmt.Errorf("database failure: %w", err)
						}
					}
				}
			}
		}

	}

	return SyncSmartCartCompatibilityCommandResult{Status: true}, nil
}
