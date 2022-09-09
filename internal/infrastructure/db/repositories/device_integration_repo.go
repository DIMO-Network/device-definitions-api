//go:generate mockgen -source device_integration_repo.go -destination mocks/device_integration_repo_mock.go -package mocks

package repositories

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DeviceIntegrationRepository interface {
	Create(ctx context.Context, deviceDefinitionID string, integrationID string, region string) (*models.DeviceIntegration, error)
}

type deviceIntegrationRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceIntegrationRepository(dbs func() *db.ReaderWriter) DeviceIntegrationRepository {
	return &deviceIntegrationRepository{DBS: dbs}
}

func (r *deviceIntegrationRepository) Create(ctx context.Context, deviceDefinitionID string, integrationID string, region string) (*models.DeviceIntegration, error) {

	di := &models.DeviceIntegration{
		DeviceDefinitionID: deviceDefinitionID,
		IntegrationID:      integrationID,
		Region:             region,
	}
	err := di.Insert(ctx, r.DBS().Writer, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return di, nil
}
