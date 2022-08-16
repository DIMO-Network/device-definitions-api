package repositories

import (
	"context"
	"database/sql"
	"errors"

	interfaces "github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/core/interfaces/repositories"
	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/infrastructure/db/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DeviceDefinitionRepository struct {
	Db *sql.DB
}

func NewDeviceDefinitionRepository(db *sql.DB) interfaces.IDeviceDefinitionRepository {
	return &DeviceDefinitionRepository{
		Db: db,
	}
}

func (r *DeviceDefinitionRepository) GetByMakeModelAndYears(ctx context.Context, make string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.Where("dm.name ilike ?", make),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(12),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	}
	if loadIntegrations {
		qms = append(qms,
			qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
			qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration)))
	}

	query := models.DeviceDefinitions(qms...)
	dd, err := query.One(ctx, r.Db)
	if err != nil {
		return nil, err
	}
	return dd, nil
}

func (r *DeviceDefinitionRepository) GetAll(ctx context.Context) ([]*models.DeviceDefinition, error) {
	dd, err := models.DeviceDefinitions().All(ctx, r.Db)
	if err != nil {
		return nil, err
	}

	return dd, err
}

func (r *DeviceDefinitionRepository) GetById(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration))).
		One(ctx, r.Db)

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

func (r *DeviceDefinitionRepository) GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		qm.Where("id = ?", id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load("DeviceIntegrations.Integration")).
		One(ctx, r.Db)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err)
		}
		return nil, nil
	}

	return dd, nil
}
