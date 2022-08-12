package repositories

import (
	"context"
	"database/sql"
	"errors"

	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
)

type DeviceDefinitionRepository struct {
	Db *sql.DB
}

func NewDeviceDefinitionRepository(db *sql.DB) interfaces.IDeviceDefinitionRepository {
	return &DeviceDefinitionRepository{
		Db: db,
	}
}

func (r *DeviceDefinitionRepository) GetById(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(models.DeviceDefinitionWhere.ID.EQ(id)).One(ctx, r.Db)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err)
		}
		return nil, nil
	}

	return dd, nil
}
