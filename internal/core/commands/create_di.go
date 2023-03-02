package commands

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/TheFellow/go-mediator/mediator"
)

type CreateDeviceIntegrationCommand struct {
	DeviceDefinitionID string                                                `json:"device_definition_id"`
	IntegrationID      string                                                `json:"integration_id"`
	Region             string                                                `json:"region"`
	Features           []*coremodels.UpdateDeviceIntegrationFeatureAttribute `json:"features"`
}

type CreateDeviceIntegrationCommandResult struct {
	ID string `json:"id"`
}

func (*CreateDeviceIntegrationCommand) Key() string { return "CreateDeviceIntegrationCommand" }

type CreateDeviceIntegrationCommandHandler struct {
	Repository                 repositories.DeviceIntegrationRepository
	DBS                        func() *db.ReaderWriter
	DDCache                    services.DeviceDefinitionCacheService
	DeviceDefinitionRepository repositories.DeviceDefinitionRepository
}

func NewCreateDeviceIntegrationCommandHandler(repository repositories.DeviceIntegrationRepository,
	dbs func() *db.ReaderWriter,
	cache services.DeviceDefinitionCacheService,
	deviceDefinitionRepository repositories.DeviceDefinitionRepository) CreateDeviceIntegrationCommandHandler {
	return CreateDeviceIntegrationCommandHandler{Repository: repository, DBS: dbs, DDCache: cache, DeviceDefinitionRepository: deviceDefinitionRepository}
}

func (ch CreateDeviceIntegrationCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	command := query.(*CreateDeviceIntegrationCommand)

	features, err := repoModel.IntegrationFeatures().All(ctx, ch.DBS().Reader)

	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get integration features"),
		}
	}

	integrationFeaturesValues, err := common.BuildDeviceIntegrationFeatureAttribute(command.Features, features)

	if err != nil {
		return nil, &exceptions.ValidationError{
			Err: err,
		}
	}

	di, err := ch.Repository.Create(ctx, command.DeviceDefinitionID, command.IntegrationID, command.Region, integrationFeaturesValues)

	if err != nil {
		return nil, err
	}

	dd, err := ch.DeviceDefinitionRepository.GetByID(ctx, command.DeviceDefinitionID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, dd.ID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dd.R.DeviceMake.Name, dd.Model, int(dd.Year))
	ch.DDCache.DeleteDeviceDefinitionCacheBySlug(ctx, dd.ModelSlug, int(dd.Year))

	return CreateDeviceIntegrationCommandResult{ID: di.IntegrationID}, nil
}
