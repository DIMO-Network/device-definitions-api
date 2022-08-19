//go:generate mockgen -source device_definition_repo.go -destination mocks/device_definition_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DeviceDefinitionRepository interface {
	GetById(ctx context.Context, id string) (*models.DeviceDefinition, error)
	GetByMakeModelAndYears(ctx context.Context, make string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	GetAll(ctx context.Context, verified bool) ([]*models.DeviceDefinition, error)
	GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error)
}

type deviceDefinitionRepository struct {
	DBS func() *db.DBReaderWriter
}

func NewDeviceDefinitionRepository(dbs func() *db.DBReaderWriter) DeviceDefinitionRepository {
	return &deviceDefinitionRepository{DBS: dbs}
}

func (r *deviceDefinitionRepository) GetByMakeModelAndYears(ctx context.Context, make string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.Where("dm.name ilike ?", make),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	}
	if loadIntegrations {
		qms = append(qms,
			qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
			qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration)))
	}

	query := models.DeviceDefinitions(qms...)
	dd, err := query.One(ctx, r.DBS().Reader)
	if err != nil {
		return nil, err
	}

	return dd, nil
}

func (r *deviceDefinitionRepository) GetAll(ctx context.Context, verified bool) ([]*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(qm.Where("verified = true"),
		qm.OrderBy("device_make_id, model, year")).All(ctx, r.DBS().Reader)

	if err != nil {
		return nil, err
	}

	return dd, err
}

func (r *deviceDefinitionRepository) GetById(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration))).
		One(ctx, r.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err)
		}
		return nil, nil
	}

	if dd.R == nil || dd.R.DeviceMake == nil {
		return nil, errors.New("required DeviceMake relation is not set")
	}

	return dd, nil
}

func (r *deviceDefinitionRepository) GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load("DeviceIntegrations.Integration")).
		One(ctx, r.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err)
		}
		return nil, nil
	}

	return dd, nil
}
