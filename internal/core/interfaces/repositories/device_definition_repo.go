//go:generate mockgen -source device_definition_repo.go -destination mocks/device_definition_repo_mock.go -package mocks
package interfaces

import (
	"context"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
)

type IDeviceDefinitionRepository interface {
	GetById(ctx context.Context, id string) (*models.DeviceDefinition, error)
}
