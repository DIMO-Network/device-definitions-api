//go:generate mockgen -source device_integration_repo.go -destination mocks/device_integration_repo_mock.go -package mocks

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
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

func NewDeviceIntegrationRepository(dbs func() *db.ReaderWriter) DeviceIntegrationRepository {
	return &deviceIntegrationRepository{DBS: dbs}
}

const (
	AutoPiVendor      = "AutoPi"
	AutoPiWebhookPath = "/webhooks/autopi-command"
)

// IntegrationsMetadata represents json stored in integrations table metadata jsonb column
type IntegrationsMetadata struct {
	AutoPiDefaultTemplateID      int                    `json:"autoPiDefaultTemplateId"`
	AutoPiPowertrainToTemplateID map[PowertrainType]int `json:"autoPiPowertrainToTemplateId,omitempty"`
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

func (p *PowertrainType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Potentially an invalid value.
	switch bv := PowertrainType(s); bv {
	case ICE, HEV, PHEV, BEV, FCEV:
		*p = bv
		return nil
	default:
		return fmt.Errorf("unrecognized value: %s", s)
	}
}

func (r *deviceIntegrationRepository) Create(ctx context.Context, deviceDefinitionID string, integrationID string, region string) (*models.DeviceIntegration, error) {

	const (
		autoPiType        = models.IntegrationTypeHardware
		autoPiStyle       = models.IntegrationStyleAddon
		defaultTemplateID = 10
	)

	integration, err := models.Integrations(models.IntegrationWhere.Vendor.EQ(AutoPiVendor),
		models.IntegrationWhere.Style.EQ(autoPiStyle), models.IntegrationWhere.Type.EQ(autoPiType)).
		One(ctx, r.DBS().Writer)

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
				return nil, errors.Wrap(err, "error inserting autoPi integration")
			}
		}
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Wrap(err, "error fetching autoPi integration from database")
	}

	if integration.ID == integrationID {
		// create device integ on the fly
		di := &models.DeviceIntegration{
			DeviceDefinitionID: deviceDefinitionID,
			IntegrationID:      integrationID,
			Region:             region,
		}
		err = di.Insert(ctx, r.DBS().Writer, boil.Infer())
		if err != nil {
			return nil, err
		}

		di.R = di.R.NewStruct()
		di.R.Integration = integration
		return di, nil
	}

	return nil, nil
}
