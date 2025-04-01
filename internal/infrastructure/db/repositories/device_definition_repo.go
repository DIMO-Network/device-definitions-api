//go:generate mockgen -source device_definition_repo.go -destination mocks/device_definition_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"strings"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type GetDeviceCompatibilityRequest struct {
	MakeID        string `json:"makeId" validate:"required"`
	IntegrationID string `json:"integrationId" validate:"required"`
	Region        string `json:"region" validate:"required"`
	Cursor        string `json:"cursor"`
	Size          int64  `json:"size"`
}

type DeviceDefinitionRepository interface {
	GetByID(ctx context.Context, id string) (*models.DeviceDefinition, error)
	GetByMakeModelAndYears(ctx context.Context, mk string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	GetBySlugAndYear(ctx context.Context, slug string, year int, loadIntegrations bool) (*models.DeviceDefinition, error)
	GetBySlugName(ctx context.Context, slug string, loadIntegrations bool) (*models.DeviceDefinition, error)
	GetAll(ctx context.Context) ([]*models.DeviceDefinition, error)
	GetDevicesByMakeYearRange(ctx context.Context, makeID string, yearStart, yearEnd int32) ([]*models.DeviceDefinition, error)
	GetDevicesMMY(ctx context.Context) ([]*DeviceMMYJoinQueryOutput, error)
	GetWithIntegrations(ctx context.Context, id string) (*models.DeviceDefinition, error)
	GetOrCreate(ctx context.Context, tx *sql.Tx, source string, extID string, makeOrID string, model string, year int, deviceTypeID string, metaData null.JSON, verified bool, hardwareTemplateID *string) (*models.DeviceDefinition, error)
	CreateOrUpdate(ctx context.Context, dd *models.DeviceDefinition, deviceStyles []*models.DeviceStyle, deviceIntegrations []*models.DeviceIntegration) (*models.DeviceDefinition, error)
}

type deviceDefinitionRepository struct {
	DBS                            func() *db.ReaderWriter
	deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService
}

func NewDeviceDefinitionRepository(dbs func() *db.ReaderWriter, deviceDefinitionOnChainService gateways.DeviceDefinitionOnChainService) DeviceDefinitionRepository {
	return &deviceDefinitionRepository{DBS: dbs, deviceDefinitionOnChainService: deviceDefinitionOnChainService}
}

func (r *deviceDefinitionRepository) GetByMakeModelAndYears(ctx context.Context, mk string, model string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_definitions_api.device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.Where("dm.name ilike ?", mk),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DefinitionImages),
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

func (r *deviceDefinitionRepository) GetBySlugAndYear(ctx context.Context, slug string, year int, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_definitions_api.device_makes dm on dm.id = device_definitions.device_make_id"),
		models.DeviceDefinitionWhere.ModelSlug.EQ(slug),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DefinitionImages),
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

func (r *deviceDefinitionRepository) GetBySlugName(ctx context.Context, slug string, loadIntegrations bool) (*models.DeviceDefinition, error) {
	qms := []qm.QueryMod{
		qm.InnerJoin("device_definitions_api.device_makes dm on dm.id = device_definitions.device_make_id"),
		models.DeviceDefinitionWhere.NameSlug.EQ(slug),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DefinitionImages),
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

func (r *deviceDefinitionRepository) GetAll(ctx context.Context) ([]*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DefinitionImages),
		models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.OrderBy("device_make_id, model, year")).All(ctx, r.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.DeviceDefinition{}, err
		}

		return nil, &exceptions.InternalError{Err: err}
	}

	return dd, err
}

func (r *deviceDefinitionRepository) GetDevicesByMakeYearRange(ctx context.Context, makeID string, yearStart, yearEnd int32) ([]*models.DeviceDefinition, error) {
	dd, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.DeviceMakeID.EQ(makeID),
		models.DeviceDefinitionWhere.Year.GTE(int16(yearStart)),
		models.DeviceDefinitionWhere.Year.LTE(int16(yearEnd)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		models.DeviceDefinitionWhere.Verified.EQ(true),
		qm.OrderBy("model, year")).All(ctx, r.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.DeviceDefinition{}, err
		}

		return nil, &exceptions.InternalError{Err: err}
	}

	return dd, err
}

type DeviceMMYJoinQueryOutput struct {
	DefinitionNameSlug string `boil:"definition_name_slug"`
	ModelSlug          string `boil:"model_slug"`
	Year               int16  `boil:"year"`
	MakeSlug           string `boil:"make_name_slug"`
}

func (r *deviceDefinitionRepository) GetDevicesMMY(ctx context.Context) ([]*DeviceMMYJoinQueryOutput, error) {
	// loads only certain properties of devices: make_slug, model_slug and year
	result := make([]*DeviceMMYJoinQueryOutput, 0)

	err := queries.Raw(
		`SELECT 	dd.name_slug as definition_name_slug, 
       					dd.model_slug, 
       					dd.year, 
       					dm.name_slug as make_name_slug 
				FROM device_definitions_api.device_definitions dd
    			INNER JOIN device_definitions_api.device_makes dm 
        		ON dm.id = dd.device_make_id`).
		Bind(ctx, r.DBS().Reader, &result)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, err
		}
		return nil, &exceptions.InternalError{Err: err}
	}

	return result, err
}

func (r *deviceDefinitionRepository) GetByID(ctx context.Context, id string) (*models.DeviceDefinition, error) {

	dd, err := models.DeviceDefinitions(
		models.DeviceDefinitionWhere.ID.EQ(id),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
		qm.Load(models.DeviceDefinitionRels.DeviceType),
		qm.Load(models.DeviceDefinitionRels.DefinitionImages),
		qm.Load(models.DeviceDefinitionRels.DefinitionDeviceStyles)).
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

func (r *deviceDefinitionRepository) GetOrCreate(ctx context.Context, tx *sql.Tx, source string, extID string, makeOrID string, model string, year int, deviceTypeID string, metaData null.JSON, verified bool, hardwareTemplateID *string) (*models.DeviceDefinition, error) {
	model = strings.TrimSpace(model)
	if len(model) == 0 {
		return nil, errors.New("invalid model, must not be blank")
	}
	commitTx := false
	if tx == nil {
		tx, _ = r.DBS().Writer.BeginTx(ctx, nil)
		commitTx = true
		defer tx.Rollback() //nolint
	}

	qms := []qm.QueryMod{
		qm.InnerJoin("device_definitions_api.device_makes dm on dm.id = device_definitions.device_make_id"),
		qm.And("model ilike ?", model),
		models.DeviceDefinitionWhere.Year.EQ(int16(year)),
		qm.Load(models.DeviceDefinitionRels.DeviceMake),
	}
	if len(makeOrID) == 27 { // i checked, no makes w/ length of 27 currently
		qms = append(qms, models.DeviceDefinitionWhere.DeviceMakeID.EQ(makeOrID))
	} else {
		qms = append(qms, qm.Where("dm.name ilike ?", makeOrID))
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

	if hardwareTemplateID == nil {
		h := common.DefautlAutoPiTemplate
		hardwareTemplateID = &h
	}

	// Create device Make
	allowCreate := true
	qmsMake := models.DeviceMakeWhere.Name.EQ(strings.TrimSpace(makeOrID))
	if len(makeOrID) == 27 {
		allowCreate = false
		qmsMake = models.DeviceMakeWhere.ID.EQ(strings.TrimSpace(makeOrID))
	}
	m, err := models.DeviceMakes(qmsMake).One(ctx, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if !allowCreate {
				return nil, &exceptions.NotFoundError{Err: fmt.Errorf("could not find makeId: %s", makeOrID)}
			}
			if len(makeOrID) == 0 {
				return nil, &exceptions.ValidationError{Err: fmt.Errorf("could not insert a blank mark")}
			}
			// create
			m = &models.DeviceMake{
				ID:       ksuid.New().String(),
				Name:     makeOrID,
				NameSlug: shared.SlugString(makeOrID),
			}
			err = m.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: errors.Wrapf(err, "error inserting make: %s", makeOrID),
				}
			}
		}
	}
	integration, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(common.AutoPiVendor)).One(ctx, r.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{Err: errors.Wrap(err, "failed to get autopi integration")}
	}

	modelSlug := shared.SlugString(model)
	nameSlug := common.DeviceDefinitionSlug(m.NameSlug, modelSlug, int16(year))

	dd = &models.DeviceDefinition{
		ID:                 ksuid.New().String(),
		DeviceMakeID:       m.ID,
		Model:              model,
		Year:               int16(year),
		Source:             null.StringFrom(source),
		Verified:           verified,
		ModelSlug:          modelSlug,
		DeviceTypeID:       null.StringFrom(deviceTypeID),
		HardwareTemplateID: null.StringFromPtr(hardwareTemplateID),
		NameSlug:           nameSlug,
	}
	// set external id's
	extIDs := []*coremodels.ExternalID{{
		Vendor: source,
		ID:     extID,
	}}
	_ = dd.ExternalIds.Marshal(extIDs)

	if metaData.Valid {
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

	// by default add autopi compatibility
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

	if commitTx {
		// Create DD onchain
		trx, err := r.deviceDefinitionOnChainService.Create(ctx, *m, *dd)
		if err != nil {
			return nil, &exceptions.InternalError{Err: err}
		}
		// add transaction info to db
		if trx != nil {
			trxArray := strings.Split(*trx, ",")
			if dd.TRXHashHex != nil {
				dd.TRXHashHex = append(dd.TRXHashHex, trxArray...)
			} else {
				dd.TRXHashHex = trxArray
			}
		}

		err = tx.Commit()
		if err != nil {
			return nil, &exceptions.InternalError{Err: err}
		}
	}

	return dd, nil
}

// CreateOrUpdate does an upsert to persist the device definition. Includes metadata as a parameter, device styles will be created on the fly
// uses a transaction to rollback if any part does not get written
func (r *deviceDefinitionRepository) CreateOrUpdate(ctx context.Context, dd *models.DeviceDefinition, deviceStyles []*models.DeviceStyle, deviceIntegrations []*models.DeviceIntegration) (*models.DeviceDefinition, error) {
	tx, _ := r.DBS().Writer.BeginTx(ctx, nil)
	defer tx.Rollback() //nolint

	if dd.HardwareTemplateID.String == "" {
		dd.HardwareTemplateID = null.StringFrom(common.DefautlAutoPiTemplate)
	}

	if err := dd.Upsert(ctx, tx, true, []string{models.DeviceDefinitionColumns.ID}, boil.Infer(), boil.Infer()); err != nil {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	if len(deviceStyles) > 0 {
		// Remove Device Styles
		_, err := models.DeviceStyles(models.DeviceStyleWhere.DefinitionID.EQ(dd.NameSlug)).
			DeleteAll(ctx, tx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}

		// Update Device Styles
		for _, ds := range deviceStyles {
			deviceStyleID := ds.ID

			if len(deviceStyleID) == 0 {
				deviceStyleID = ksuid.New().String()
			}

			subModels := &models.DeviceStyle{
				ID:                 deviceStyleID,
				DefinitionID:       dd.NameSlug,
				Name:               ds.Name,
				ExternalStyleID:    ds.ExternalStyleID,
				Source:             ds.Source,
				CreatedAt:          ds.CreatedAt,
				UpdatedAt:          ds.UpdatedAt,
				SubModel:           ds.SubModel,
				HardwareTemplateID: ds.HardwareTemplateID,
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
				CreatedAt:          di.CreatedAt,
				UpdatedAt:          di.UpdatedAt,
				Region:             di.Region,
				Features:           di.Features,
			}
			err = deviceIntegration.Insert(ctx, tx, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: err,
				}
			}
		}
	}

	err := tx.Commit()
	if err != nil {
		return nil, &exceptions.InternalError{Err: err}
	}

	// Create onchain
	if dd.R != nil && dd.R.DeviceMake != nil {
		_, err := r.deviceDefinitionOnChainService.Create(ctx, *dd.R.DeviceMake, *dd)
		if err != nil {
			return nil, &exceptions.InternalError{Err: err}
		}
	}

	return dd, nil
}
