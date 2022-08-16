//go:generate mockgen -source device_make_repo.go -destination mocks/device_make_repo_mock.go -package mocks
package interfaces

import (
	"context"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
)

type IDeviceMakeRepository interface {
	GetAll(ctx context.Context) ([]*models.DeviceMake, error)
}
