//nolint:tagliatelle
package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/repositories"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

type UpdateDeviceDefinitionCommand struct {
	DeviceDefinitionID string                      `json:"deviceDefinitionId"`
	Source             string                      `json:"source"`
	ExternalID         string                      `json:"external_id"`
	ImageURL           string                      `json:"image_url"`
	Verified           bool                        `json:"verified"`
	Model              string                      `json:"model"`
	Year               int16                       `json:"year"`
	HardwareTemplateID string                      `json:"hardware_template_id,omitempty"`
	DeviceMakeID       string                      `json:"device_make_id"`
	DeviceStyles       []*UpdateDeviceStyles       `json:"deviceStyles"`
	DeviceIntegrations []*UpdateDeviceIntegrations `json:"deviceIntegrations"`
	// DeviceTypeID comes from the device_types.id table, determines what kind of device this is, typically a vehicle
	DeviceTypeID string `json:"device_type_id"`
	// DeviceAttributes sets definition metadata eg. vehicle info. Allowed key/values are defined in device_types.properties
	DeviceAttributes []*coremodels.UpdateDeviceTypeAttribute `json:"deviceAttributes"`
	ExternalIDs      []*coremodels.ExternalID                `json:"externalIds"`
}

type UpdateDeviceIntegrations struct {
	IntegrationID string                                                `json:"integration_id"`
	Capabilities  null.JSON                                             `json:"capabilities,omitempty"`
	CreatedAt     time.Time                                             `json:"created_at,omitempty"`
	UpdatedAt     time.Time                                             `json:"updated_at,omitempty"`
	Region        string                                                `json:"region"`
	Features      []*coremodels.UpdateDeviceIntegrationFeatureAttribute `json:"features"`
}

type UpdateDeviceStyles struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ExternalStyleID    string    `json:"external_style_id"`
	Source             string    `json:"source"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	UpdatedAt          time.Time `json:"updated_at,omitempty"`
	SubModel           string    `json:"sub_model"`
	HardwareTemplateID string    `json:"hardware_template_id"`
}

type UpdateDeviceDefinitionCommandResult struct {
	ID string `json:"id"`
}

func (*UpdateDeviceDefinitionCommand) Key() string { return "UpdateDeviceDefinitionCommand" }

type UpdateDeviceDefinitionCommandHandler struct {
	Repository repositories.DeviceDefinitionRepository
	DBS        func() *db.ReaderWriter
	DDCache    services.DeviceDefinitionCacheService
}

func NewUpdateDeviceDefinitionCommandHandler(repository repositories.DeviceDefinitionRepository, dbs func() *db.ReaderWriter, cache services.DeviceDefinitionCacheService) UpdateDeviceDefinitionCommandHandler {
	return UpdateDeviceDefinitionCommandHandler{DDCache: cache, Repository: repository, DBS: dbs}
}

// Handle will update an existing device def, or if it doesn't exist create it on the fly. We may want to change create to be explicit
func (ch UpdateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {

	command := query.(*UpdateDeviceDefinitionCommand)

	if len(command.DeviceTypeID) == 0 {
		command.DeviceTypeID = common.DefaultDeviceType
	}
	if err := command.ValidateUpdate(); err != nil {
		return nil, &exceptions.ValidationError{
			Err: errors.Wrap(err, "failed model validation"),
		}
	}
	// future: either rename method to be CreateOrUpdate, and remove Create method, or only allow Updating in this method
	dd, err := ch.Repository.GetByID(ctx, command.DeviceDefinitionID)
	if err != nil {
		// if dd is not found, we'll try to create it
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.InternalError{
				Err: err,
			}
		}
	}
	// creates if does not exist
	if dd == nil {
		dd = &models.DeviceDefinition{
			ID:                 command.DeviceDefinitionID,
			DeviceMakeID:       command.DeviceMakeID,
			Model:              command.Model,
			Year:               command.Year,
			ModelSlug:          shared.SlugString(command.Model),
			HardwareTemplateID: null.StringFrom(command.HardwareTemplateID),
		}
	}

	if len(command.HardwareTemplateID) > 0 {
		dd.HardwareTemplateID = null.StringFrom(command.HardwareTemplateID)
	}

	if len(command.Model) > 0 {
		dd.Model = command.Model
	}

	if command.Year > 0 {
		dd.Year = command.Year
	}

	if len(command.DeviceMakeID) > 0 {
		dd.DeviceMakeID = command.DeviceMakeID
	}

	// Resolve make, used later to clear cache. here just as way to make sure id exists
	dm, err := models.DeviceMakes(models.DeviceMakeWhere.ID.EQ(dd.DeviceMakeID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes with make id: %s", command.DeviceMakeID),
		}
	}

	// Resolve attributes by device types
	dt, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.DeviceTypeID)).One(ctx, ch.DBS().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.ValidationError{
				Err: fmt.Errorf("device type id: %s not found when updating a definition", command.DeviceTypeID),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device types"),
		}
	}
	dd.DeviceTypeID = null.StringFrom(dt.ID)

	// attribute info - deviceTypeInfo will be json invalid if command.DeviceAttributes is nil
	deviceTypeInfo, err := common.BuildDeviceTypeAttributes(command.DeviceAttributes, dt)
	if err != nil {
		return nil, err
	}
	if deviceTypeInfo.Valid {
		dd.Metadata = deviceTypeInfo
	}

	if len(command.ExternalID) > 0 {
		dd.ExternalID = null.StringFrom(command.ExternalID)
	}

	extIDs := map[string]string{}
	if len(command.Source) > 0 {
		dd.Source = null.StringFrom(command.Source)
		dd.ExternalID = null.StringFrom(command.ExternalID)
		extIDs[command.Source] = command.ExternalID
	}
	if len(command.ExternalIDs) > 0 {
		for _, ei := range command.ExternalIDs {
			extIDs[ei.Vendor] = ei.ID
		}
	}
	extIDsJSON, err := json.Marshal(extIDs)
	if err != nil {
		return nil, err
	}
	dd.ExternalIds = null.JSONFrom(extIDsJSON)

	// todo if the command has an image, let's determine how that should be handled, or remove it from the command

	// if a definition was previously marked as verified, we do not want to go back and un-verify it, at least not in this flow. This will only mark DD's as verified
	if command.Verified {
		dd.Verified = command.Verified // tech debt, there may be real case where we want to un-verify eg. admin tool
	}
	var deviceStyles []*models.DeviceStyle
	var deviceIntegrations []*models.DeviceIntegration

	if len(command.DeviceStyles) > 0 {
		for _, ds := range command.DeviceStyles {
			deviceStyles = append(deviceStyles, &models.DeviceStyle{
				ID:                 ds.ID,
				DeviceDefinitionID: command.DeviceDefinitionID,
				Name:               ds.Name,
				ExternalStyleID:    ds.ExternalStyleID,
				Source:             ds.Source,
				CreatedAt:          ds.CreatedAt,
				UpdatedAt:          ds.UpdatedAt,
				SubModel:           ds.SubModel,
				HardwareTemplateID: null.StringFrom(ds.HardwareTemplateID),
			})
		}
	}

	if len(command.DeviceIntegrations) > 0 {
		features, err := models.IntegrationFeatures().All(ctx, ch.DBS().Reader)
		if err != nil {
			return nil, &exceptions.InternalError{
				Err: fmt.Errorf("failed to get integration features"),
			}
		}
		for _, di := range command.DeviceIntegrations {
			deviceIntegration := &models.DeviceIntegration{
				DeviceDefinitionID: command.DeviceDefinitionID,
				IntegrationID:      di.IntegrationID,
				CreatedAt:          di.CreatedAt,
				UpdatedAt:          di.UpdatedAt,
				Region:             di.Region,
			}

			integrationFeaturesValues, err := common.BuildDeviceIntegrationFeatureAttribute(di.Features, features)
			if err != nil {
				return nil, &exceptions.ValidationError{
					Err: err,
				}
			}

			err = deviceIntegration.Features.Marshal(integrationFeaturesValues)
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: err,
				}
			}
			deviceIntegrations = append(deviceIntegrations, deviceIntegration)
		}
	}

	// if deviceTypeInfo is nil, no metadata will be updated
	dd, err = ch.Repository.CreateOrUpdate(ctx, dd, deviceStyles, deviceIntegrations)

	if err != nil {
		return nil, err
	}

	// Remove Cache
	ch.DDCache.DeleteDeviceDefinitionCacheByID(ctx, dd.ID)
	ch.DDCache.DeleteDeviceDefinitionCacheByMakeModelAndYears(ctx, dm.Name, dd.Model, int(dd.Year))
	ch.DDCache.DeleteDeviceDefinitionCacheBySlug(ctx, dd.ModelSlug, int(dd.Year))

	return UpdateDeviceDefinitionCommandResult{ID: dd.ID}, nil
}

// ValidateUpdate validates the contents of a UpdateDeviceDefinitionCommand for purpose of updating record
func (udc *UpdateDeviceDefinitionCommand) ValidateUpdate() error {
	return validation.ValidateStruct(udc,
		validation.Field(&udc.DeviceDefinitionID, validation.Required),
		validation.Field(&udc.DeviceDefinitionID, validation.Length(27, 27)),
	)
}
