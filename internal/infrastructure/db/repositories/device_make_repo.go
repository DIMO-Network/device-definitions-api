package repositories

import (
	"context"
	"database/sql"

	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
)

type DeviceMakeRepository struct {
	Db *sql.DB
}

func NewDeviceMakeRepository(db *sql.DB) interfaces.IDeviceMakeRepository {
	return &DeviceMakeRepository{
		Db: db,
	}
}

func (r *DeviceMakeRepository) GetAll(ctx context.Context) ([]*models.DeviceMake, error) {
	makes, err := models.DeviceMakes().All(ctx, r.Db)
	if err != nil {
		return nil, err
	}

	return makes, err
}
