//go:generate mockgen -source vin_repo.go -destination mocks/vin_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	stringutils "github.com/DIMO-Network/shared/pkg/strings"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type VINRepository interface {
	GetOrCreateWMI(ctx context.Context, wmi string, mk string) (*models.Wmi, error)
}

type vinRepository struct {
	DBS              func() *db.ReaderWriter
	registryInstance *contracts.Registry
	identity         gateways.IdentityAPI
}

func NewVINRepository(dbs func() *db.ReaderWriter, registryInstance *contracts.Registry, identity gateways.IdentityAPI) VINRepository {
	return &vinRepository{DBS: dbs, registryInstance: registryInstance, identity: identity}
}

func (r *vinRepository) GetOrCreateWMI(ctx context.Context, wmi string, mk string) (*models.Wmi, error) {
	if len(wmi) != 3 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid wmi for GetOrCreate: %s", wmi)}
	}
	if len(mk) < 2 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid make name for GetOrCreate: %s", mk)}
	}
	makeSlug := stringutils.SlugString(mk)

	deviceMake, err := r.identity.GetManufacturer(makeSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{
				Err: errors.Errorf("failed to find makeSlug from vin decode with name slug: %s", makeSlug),
			}
		}
		return nil, err
	}
	manufacturerID, err := r.registryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, deviceMake.Name)
	if err != nil || manufacturerID == nil {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("make has not been minted yet or no tokenID set: %s :%s", mk, err)}
	}

	//dbWMI, err := models.FindWmi(ctx, r.dbs().Reader, wmi, deviceMake.ID) // there can be WMI's for more than one Make
	dbWMI, err := models.Wmis(models.WmiWhere.Wmi.EQ(wmi), models.WmiWhere.ManufacturerName.EQ(mk)).
		One(ctx, r.DBS().Reader) // there can be WMI's for more than one Make
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, &exceptions.InternalError{
			Err: err,
		}
	}

	if dbWMI == nil {
		dbWMI = &models.Wmi{
			Wmi:              wmi,
			ManufacturerName: mk,
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
