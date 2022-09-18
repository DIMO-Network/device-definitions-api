//go:generate mockgen -source device_integration_repo.go -destination mocks/device_integration_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DeviceIntegrationRepository interface {
	Create(ctx context.Context, deviceDefinitionID string, integrationID string, region string) (*models.DeviceIntegration, error)
}

type deviceIntegrationRepository struct {
	DBS func() *db.ReaderWriter
}

type PowertrainType string

const (
	ICE  PowertrainType = "ICE"
	HEV  PowertrainType = "HEV"
	PHEV PowertrainType = "PHEV"
	BEV  PowertrainType = "BEV"
	FCEV PowertrainType = "FCEV"
)

func (p PowertrainType) String() string {
	return string(p)
}

// IntegrationsMetadata represents json stored in integrations table metadata jsonb column
type IntegrationsMetadata struct {
	AutoPiDefaultTemplateID      int                    `json:"autoPiDefaultTemplateId"`
	AutoPiPowertrainToTemplateID map[PowertrainType]int `json:"autoPiPowertrainToTemplateId,omitempty"`
}

func NewDeviceIntegrationRepository(dbs func() *db.ReaderWriter) DeviceIntegrationRepository {
	return &deviceIntegrationRepository{DBS: dbs}
}

func (r *deviceIntegrationRepository) Create(ctx context.Context, deviceDefinitionID string, integrationID string, region string) (*models.DeviceIntegration, error) {

	const (
		AutoPiVendor      = "AutoPi"
		autoPiType        = models.IntegrationTypeHardware
		autoPiStyle       = models.IntegrationStyleAddon
		defaultTemplateID = 10
	)

	// Validate if the integrationID was sent.
	if len(integrationID) > 0 {
		integration, err := models.Integrations(models.IntegrationWhere.ID.EQ(integrationID)).
			One(ctx, r.DBS().Reader)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, &exceptions.NotFoundError{
					Err: fmt.Errorf("could not find integration id: %s", integrationID),
				}
			}
		}

		integrationID = integration.ID
	} else {
		integration, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(AutoPiVendor),
			models.IntegrationWhere.Style.EQ(autoPiStyle), models.IntegrationWhere.Type.EQ(autoPiType)).
			One(ctx, r.DBS().Reader)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// create
				im := IntegrationsMetadata{AutoPiDefaultTemplateID: defaultTemplateID}
				integration = &models.Integration{
					ID:     ksuid.New().String(),
					Vendor: AutoPiVendor,
					Type:   autoPiType,
					Style:  autoPiStyle,
				}
				_ = integration.Metadata.Marshal(im)
				err = integration.Insert(ctx, r.DBS().Writer, boil.Infer())
				if err != nil {
					return nil, &exceptions.InternalError{
						Err: errors.Wrap(err, "error inserting autoPi integration"),
					}
				}
			} else {
				return nil, &exceptions.InternalError{
					Err: errors.Wrap(err, "error fetching autoPi integration from database"),
				}
			}
		}

		integrationID = integration.ID
	}

	di, err := models.DeviceIntegrations(models.DeviceIntegrationWhere.DeviceDefinitionID.EQ(deviceDefinitionID),
		models.DeviceIntegrationWhere.IntegrationID.EQ(integrationID), models.DeviceIntegrationWhere.Region.EQ(region)).
		One(ctx, r.DBS().Reader)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	if di == nil {
		di = &models.DeviceIntegration{
			DeviceDefinitionID: deviceDefinitionID,
			IntegrationID:      integrationID,
			Region:             region,
		}
		err = di.Insert(ctx, r.DBS().Writer, boil.Infer())
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}

	return di, nil
}
