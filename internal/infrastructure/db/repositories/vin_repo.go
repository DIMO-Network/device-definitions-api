//go:generate mockgen -source vin_repo.go -destination mocks/vin_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type VINRepository interface {
	GetOrCreateWMI(ctx context.Context, wmi string, make string) (*models.Wmi, error)
}

type vinRepository struct {
	DBS func() *db.ReaderWriter
}

func NewVINRepository(dbs func() *db.ReaderWriter) VINRepository {
	return &vinRepository{DBS: dbs}
}

func (r *vinRepository) GetOrCreateWMI(ctx context.Context, wmi string, make string) (*models.Wmi, error) {
	if len(wmi) != 3 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid wmi for GetOrCreate: %s", wmi)}
	}

	dbWMI, err := models.FindWmi(ctx, r.DBS().Reader, wmi)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	if dbWMI == nil {
		deviceMake, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(common.SlugString(make))).One(ctx, r.DBS().Reader)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, &exceptions.NotFoundError{
					Err: errors.Errorf("failed to find make from vin decode with name slug: %s", common.SlugString(make)),
				}
			}
		}

		dbWMI = &models.Wmi{
			Wmi:          wmi,
			DeviceMakeID: deviceMake.ID,
		}
		err = dbWMI.Insert(ctx, r.DBS().Writer, boil.Infer())
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	return dbWMI, nil
}
