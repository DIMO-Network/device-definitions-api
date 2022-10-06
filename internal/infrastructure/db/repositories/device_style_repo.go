//go:generate mockgen -source device_integration_repo.go -destination mocks/device_integration_repo_mock.go -package mocks

package repositories

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DeviceStyleRepository interface {
	Create(ctx context.Context, deviceDefinitionID string, name string, externalStyleID string, source string, subModel string) (*models.DeviceStyle, error)
}

type deviceStyleRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceStyleRepository(dbs func() *db.ReaderWriter) DeviceStyleRepository {
	return &deviceStyleRepository{DBS: dbs}
}

func (r *deviceStyleRepository) Create(ctx context.Context, deviceDefinitionID string, name string, externalStyleID string, source string, subModel string) (*models.DeviceStyle, error) {

	ds := &models.DeviceStyle{
		ID:                 ksuid.New().String(),
		DeviceDefinitionID: deviceDefinitionID,
		Name:               name,
		ExternalStyleID:    externalStyleID,
		Source:             source,
		SubModel:           subModel,
	}
	err := ds.Insert(ctx, r.DBS().Writer, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return ds, nil
}
