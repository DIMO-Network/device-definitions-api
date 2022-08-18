//go:generate mockgen -source device_make_repo.go -destination mocks/device_make_repo_mock.go -package mocks

package repositories

import (
	"context"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DeviceMakeRepository interface {
	GetAll(ctx context.Context) ([]*models.DeviceMake, error)
}

type deviceMakeRepository struct {
	DBS func() *db.DBReaderWriter
}

func NewDeviceMakeRepository(dbs func() *db.DBReaderWriter) DeviceMakeRepository {
	return &deviceMakeRepository{
		DBS: dbs,
	}
}

func (r *deviceMakeRepository) GetAll(ctx context.Context) ([]*models.DeviceMake, error) {
	makes, err := models.DeviceMakes(qm.OrderBy(models.DeviceMakeColumns.Name)).All(ctx, r.DBS().Reader)

	if err != nil {
		return nil, err
	}

	return makes, err
}
