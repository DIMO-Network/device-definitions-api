//go:generate mockgen -source device_definition_repo.go -destination mocks/device_definition_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DeviceDefinitionRepository interface {
	GetByID(ctx context.Context, id string) (*models.DeviceDefinition, error)
	GetByMakeModelAndYears(ctx context.Context, make string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	GetAll(ctx context.Context, verified bool) ([]*models.DeviceDefinition, error)
	GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error)
	GetOrCreate(ctx context.Context, source string, make string, model string, year int, deviceTypeID string, metaData map[string]interface{}) (*models.DeviceDefinition, error)
	CreateOrUpdate(ctx context.Context, dd *models.DeviceDefinition, deviceStyles []*models.DeviceStyle, deviceIntegrations []*models.DeviceIntegration, metaData map[string]interface{}) (*models.DeviceDefinition, error)
	FetchDeviceCompatibility(ctx context.Context, makeID, integrationID, region string) (models.DeviceDefinitionSlice, error)
}

type deviceDefinitionRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceDefinitionRepository(dbs func() *db.ReaderWriter) DeviceDefinitionRepository {
	return &deviceDefinitionRepository{DBS: dbs}
}

func (r *deviceDefinitionRepository) GetByMakeModelAndYears(ctx context.Context, make string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_definitions_api.device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.Where("dm.name ilike ?", make),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
	}
	if loadIntegrations {
		qms = append(qms,
			qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
			qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration)))
	}

	query := models.DeviceDefinitions(qms...)
	dd, err := query.One(ctx, r.DBS().Reader)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{Err: err}
		}

		return nil, nil
	}

	return dd, nil
}

func (r *deviceDefinitionRepository) GetAll(ctx context.Context, verified bool) ([]*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(qm.Where("verified = true"),
		qm.OrderBy("device_make_id, model, year")).All(ctx, r.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.DeviceDefinition{}, err
		}

		return nil, &exceptions.InternalError{Err: err}
	}

	return dd, err
}

func (r *deviceDefinitionRepository) GetByID(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.ID.EQ(id),
		qm.Load(models.DeviceDefinitionRels.DeviceIntegrations),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(qm.Rels(models.DeviceDefinitionRels.DeviceIntegrations, models.DeviceIntegrationRels.Integration)),
		qm.Load(models.DeviceDefinitionRels.DeviceStyles)).
		One(ctx, r.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{Err: err}
		}

		return nil, nil
	}

	if dd.R == nil || dd.R.DeviceMake == nil {
		return nil, &exceptions.ConflictError{Err: errors.New("required DeviceMake relation is not set")}
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
			return nil, &exceptions.InternalError{Err: err}
		}
		return nil, nil
	}

	return dd, nil
}

func (r *deviceDefinitionRepository) GetOrCreate(ctx context.Context, source string, make string, model string, year int, deviceTypeID string, metaData map[string]interface{}) (*models.DeviceDefinition, error) {
	tx, _ := r.DBS().Writer.BeginTx(ctx, nil)
	defer tx.Rollback() //nolint

	qms := []qm.QueryMod{
		qm.InnerJoin("device_definitions_api.device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.Where("dm.name ilike ?", make),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	}

	query := models.DeviceDefinitions(qms...)
	dd, err := query.One(ctx, r.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{Err: err}
		}
	}

	if dd != nil {
		return dd, nil
	}

	// Create device Make
	m, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(strings.TrimSpace(make))).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// create
			m = &models.DeviceMake{
				ID:       ksuid.New().String(),
				Name:     make,
				NameSlug: common.SlugString(make),
			}
			err = m.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: errors.Wrapf(err, "error inserting make: %s", make),
				}
			}
		}
	}
	integration, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(common.AutoPiVendor)).One(ctx, r.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrap(err, "failed to get autopi integration")}
	}

	dd = &models.DeviceDefinition{
		ID:           ksuid.New().String(),
		DeviceMakeID: m.ID,
		Model:        model,
		Year:         int16(year),
		Source:       null.StringFrom(source),
		Verified:     false,
		ModelSlug:    common.SlugString(model),
		DeviceTypeID: null.StringFrom(deviceTypeID),
	}

	if metaData != nil {
		err = dd.Metadata.Marshal(metaData)
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	err = dd.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}
	di := &models.DeviceIntegration{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      integration.ID,
		Region:             common.AmericasRegion.String(),
	}
	di2 := &models.DeviceIntegration{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      integration.ID,
		Region:             common.EuropeRegion.String(),
	}
	err = di.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}
	err = di2.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}

	err = tx.Commit()
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}
	return dd, nil
}

func (r *deviceDefinitionRepository) CreateOrUpdate(ctx context.Context, dd *models.DeviceDefinition, deviceStyles []*models.DeviceStyle, deviceIntegrations []*models.DeviceIntegration, metaData map[string]interface{}) (*models.DeviceDefinition, error) {

	tx, _ := r.DBS().Writer.BeginTx(ctx, nil)
	defer tx.Rollback() //nolint

	if metaData != nil {
		err := dd.Metadata.Marshal(metaData)
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if err := dd.Upsert(ctx, tx, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	if len(deviceStyles) > 0 {
		// Remove Device Styles
		_, err := models.DeviceStyles(models.DeviceStyleWhere.DeviceDefinitionID.EQ(dd.ID)).
			DeleteAll(ctx, tx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		// Update Device Styles
		for _, ds := range deviceStyles {
			subModels := &models.DeviceStyle{
				ID:                 ds.ID,
				DeviceDefinitionID: dd.ID,
				Name:               ds.Name,
				ExternalStyleID:    ds.ExternalStyleID,
				Source:             ds.Source,
				CreatedAt:          ds.CreatedAt,
				UpdatedAt:          ds.UpdatedAt,
				SubModel:           ds.SubModel,
			}
			err = subModels.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: err,
				}
			}
		}
	}

	if len(deviceIntegrations) > 0 {
		// Remove Device Integrations
		_, err := models.DeviceIntegrations(models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(dd.ID)).
			DeleteAll(ctx, tx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		for _, di := range deviceIntegrations {
			deviceIntegration := &models.DeviceIntegration{
				DeviceDefinitionID: dd.ID,
				IntegrationID:      di.IntegrationID,
				//Capabilities:       di.Capabilities,
				CreatedAt: di.CreatedAt,
				UpdatedAt: di.UpdatedAt,
				Region:    di.Region,
			}
			err = deviceIntegration.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: err,
				}
			}
		}
	}

	return dd, nil
}

func (r *deviceDefinitionRepository) FetchDeviceCompatibility(ctx context.Context, makeID, integrationID, region string) (models.DeviceDefinitionSlice, error) {
	res, err := models.DeviceDefinitions(
		qm.InnerJoin(
			models.TableNames.DeviceIntegrations+" ON "+models.DeviceDefinitionTableColumns.ID+" = "+models.DeviceIntegrationTableColumns.DeviceDefinitionID,
		),
		models.DeviceDefinitionWhere.DeviceMakeID.EQ(makeID),
		models.DeviceDefinitionWhere.Year.GTE(2008),
		models.DeviceIntegrationWhere.Features.IsNotNull(),
		models.DeviceIntegrationWhere.IntegrationID.EQ(integrationID),
		models.DeviceIntegrationWhere.Region.EQ(region),
		qm.Load(
			models.DeviceDefinitionRels.DeviceIntegrations,
			models.DeviceIntegrationWhere.IntegrationID.EQ(integrationID),
			models.DeviceIntegrationWhere.Region.EQ(region),
			models.DeviceIntegrationWhere.Features.IsNotNull(),
		),
	).All(ctx, r.DBS().Reader)

	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}

	return res, nil
}
