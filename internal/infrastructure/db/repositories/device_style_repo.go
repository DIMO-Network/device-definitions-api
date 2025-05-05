//go:generate mockgen -source device_style_repo.go -destination mocks/device_style_repo_mock.go -package mocks

package repositories

import (
	"context"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DeviceStyleRepository interface {
	Create(ctx context.Context, deviceDefinitionID string, name string, externalStyleID string, source string, subModel string, templateID string) (*models.DeviceStyle, error)
}

type deviceStyleRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceStyleRepository(dbs func() *db.ReaderWriter) DeviceStyleRepository {
	return &deviceStyleRepository{DBS: dbs}
}

func (r *deviceStyleRepository) Create(ctx context.Context, definitionID string, name string, externalStyleID string, source string, subModel string, templateID string) (*models.DeviceStyle, error) {

	ds := &models.DeviceStyle{
		ID:                 ksuid.New().String(),
		DefinitionID:       definitionID,
		Name:               name,
		ExternalStyleID:    externalStyleID,
		Source:             source,
		SubModel:           subModel,
		HardwareTemplateID: null.StringFrom(templateID),
	}
	err := ds.Insert(ctx, r.DBS().Writer, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	return ds, nil
}
