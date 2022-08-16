//go:generate mockgen -source device_definition_repo.go -destination mocks/device_definition_repo_mock.go -package mocks
package interfaces

import (
	"context"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
)

type IDeviceDefinitionRepository interface {
	GetById(ctx context.Context, id string) (*models.DeviceDefinition, error)
	GetByMakeModelAndYears(ctx context.Context, make string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	GetAll(ctx context.Context) ([]*models.DeviceDefinition, error)
	GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error)
}
