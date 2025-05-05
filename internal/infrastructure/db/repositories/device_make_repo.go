//go:generate mockgen -source device_make_repo.go -destination mocks/device_make_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"strings"

	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type DeviceMakeRepository interface {
	GetAll(ctx context.Context) ([]*models.DeviceMake, error)
	GetBySlug(ctx context.Context, slug string) (*models.DeviceMake, error)
	GetOrCreate(ctx context.Context, makeName string, logURL string, externalIDs string, metadata string, hardwareTemplateID string) (*models.DeviceMake, error)
}

type deviceMakeRepository struct {
	DBS func() *db.ReaderWriter
}

func NewDeviceMakeRepository(dbs func() *db.ReaderWriter) DeviceMakeRepository {
	return &deviceMakeRepository{
		DBS: dbs,
	}
}

func (r *deviceMakeRepository) GetAll(ctx context.Context) ([]*models.DeviceMake, error) {
	makes, err := models.DeviceMakes(qm.OrderBy(models.DeviceMakeColumns.Name)).All(ctx, r.DBS().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.DeviceMake{}, err
		}

		return nil, &exceptions.InternalError{Err: err}
	}

	return makes, err
}

func (r *deviceMakeRepository) GetOrCreate(ctx context.Context, makeName string, logURL string, externalIDs string, metadata string, hardwareTemplateID string) (*models.DeviceMake, error) {
	m, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(strings.TrimSpace(makeName))).One(ctx, r.DBS().Writer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// create
			m = &models.DeviceMake{
				ID:                 ksuid.New().String(),
				Name:               makeName,
				NameSlug:           stringutils.SlugString(makeName),
				LogoURL:            null.StringFrom(logURL),
				ExternalIds:        null.JSONFrom([]byte(externalIDs)),
				Metadata:           null.JSONFrom([]byte(metadata)),
				HardwareTemplateID: null.StringFrom(hardwareTemplateID),
			}
			// todo set TokenID. increase from latest tokenId
			err = m.Insert(ctx, r.DBS().Writer.DB, boil.Infer())
			if err != nil {
				return nil, &exceptions.InternalError{Err: errors.Wrapf(err, "error inserting make: %s", makeName)}
			}
			return m, nil
		}
		return nil, errors.Wrapf(err, "error querying for make: %s", makeName)
	}
	return m, nil
}

func (r *deviceMakeRepository) GetBySlug(ctx context.Context, slug string) (*models.DeviceMake, error) {
	m, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(strings.TrimSpace(slug))).One(ctx, r.DBS().Writer)
	if err != nil {
		return nil, errors.Wrapf(err, "error querying for make: %s", slug)
	}
	return m, nil
}
