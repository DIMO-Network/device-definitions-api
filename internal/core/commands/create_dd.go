//nolint:tagliatelle
package commands

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/DIMO-Network/shared"
	"github.com/volatiletech/null/v8"

	"github.com/DIMO-Network/device-definitions-api/internal/core/services"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/pkg/errors"
)

type CreateDeviceDefinitionCommand struct {
	Source             string `json:"source"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	Year               int    `json:"year"`
	HardwareTemplateID string `json:"hardware_template_id,omitempty"`
	// DeviceTypeID comes from the device_types.id table, determines what kind of device this is, typically a vehicle
	DeviceTypeID string `json:"device_type_id"`
	// DeviceAttributes sets definition metadata eg. vehicle info. Allowed key/values are defined in device_types.properties
	DeviceAttributes []*coremodels.UpdateDeviceTypeAttribute `json:"deviceAttributes"`
	Verified         bool                                    `json:"verified"`
}

type CreateDeviceDefinitionCommandResult struct {
	ID            string  `json:"id"`
	NameSlug      string  `json:"name_slug"`
	TransactionID *string `json:"transaction_id"`
}

func (*CreateDeviceDefinitionCommand) Key() string { return "CreateDeviceDefinitionCommand" }

type CreateDeviceDefinitionCommandHandler struct {
	onChainSvc            gateways.DeviceDefinitionOnChainService
	dbs                   func() *db.ReaderWriter
	powerTrainTypeService services.PowerTrainTypeService
	fuelAPI               gateways.FuelAPIService
	logger                *zerolog.Logger
}

func NewCreateDeviceDefinitionCommandHandler(onChainSvc gateways.DeviceDefinitionOnChainService, dbs func() *db.ReaderWriter,
	powerTrainTypeService services.PowerTrainTypeService, fuelAPI gateways.FuelAPIService, logger *zerolog.Logger) CreateDeviceDefinitionCommandHandler {
	return CreateDeviceDefinitionCommandHandler{onChainSvc: onChainSvc, dbs: dbs,
		powerTrainTypeService: powerTrainTypeService,
		fuelAPI:               fuelAPI, logger: logger}
}

func (ch CreateDeviceDefinitionCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	command := query.(*CreateDeviceDefinitionCommand)
	// DeviceTypeID is either vehicle or aftermarket_device, or as read from device_types db.
	if len(command.DeviceTypeID) == 0 {
		command.DeviceTypeID = common.DefaultDeviceType // the default is the "vehicle" name not the KSUID
	}

	// Resolve attributes by device types
	_, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(command.DeviceTypeID)).One(ctx, ch.dbs().Reader)
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

	dm, err := models.DeviceMakes(models.DeviceMakeWhere.NameSlug.EQ(shared.SlugString(command.Make))).One(ctx, ch.dbs().Reader)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.ValidationError{
				Err: fmt.Errorf("make: %s not found when updating a definition", command.Make),
			}
		}
		return nil, &exceptions.InternalError{
			Err: fmt.Errorf("failed to get device makes"),
		}
	}

	// Validate if powertraintype exists, but only if vehicle device type
	if command.DeviceTypeID == common.DefaultDeviceType {
		powerTrainExists := false
		if len(command.DeviceAttributes) > 0 {
			for _, item := range command.DeviceAttributes {
				if item.Name == common.PowerTrainType && len(item.Value) > 0 {
					powerTrainExists = true
					break
				}
			}
		}
		if !powerTrainExists {
			powerTrainTypeValue, err := ch.powerTrainTypeService.ResolvePowerTrainType(shared.SlugString(command.Make), shared.SlugString(command.Model), null.JSON{}, null.JSON{})
			if err != nil {
				return nil, &exceptions.InternalError{
					Err: fmt.Errorf("failed to get powertraintype"),
				}
			}
			if powerTrainTypeValue != "" {
				command.DeviceAttributes = append(command.DeviceAttributes, &coremodels.UpdateDeviceTypeAttribute{
					Name:  common.PowerTrainType,
					Value: powerTrainTypeValue,
				})
			}
		}
	}
	// resolve image with fuel
	images, err := ch.fuelAPI.FetchDeviceImages(command.Make, command.Model, command.Year, 2, 2)
	if err != nil {
		ch.logger.Warn().Err(err).Msgf("failed to get images for: %s %d %s", command.Make, command.Year, command.Model)
	}

	ddTbl := gateways.DeviceDefinitionTablelandModel{
		ID:         common.DeviceDefinitionSlug(dm.NameSlug, shared.SlugString(command.Model), int16(command.Year)),
		KSUID:      ksuid.New().String(),
		Model:      command.Model,
		Year:       command.Year,
		DeviceType: command.DeviceTypeID,
		ImageURI:   getDefaultImage(images),
		Metadata:   common.ConvertDeviceTypeAttrsToDefinitionMetadata(command.DeviceAttributes),
	}

	create, err := ch.onChainSvc.Create(ctx, *dm, ddTbl)
	if err != nil {
		return nil, err // todo does mediator eat this error?
	}
	err = ch.associateImagesToDeviceDefinition(ctx, ddTbl.ID, images)
	if err != nil {
		ch.logger.Err(err).Msgf("failed to add images to database for: %s %d %s", command.Make, command.Year, command.Model)
	}

	return CreateDeviceDefinitionCommandResult{ID: ddTbl.ID, NameSlug: ddTbl.ID, TransactionID: create}, nil
}

func (ch CreateDeviceDefinitionCommandHandler) associateImagesToDeviceDefinition(ctx context.Context, definitionID string, img gateways.FuelDeviceImages) error {
	var p models.Image
	// loop through all img (color variations)
	for _, device := range img.Images {
		p.ID = ksuid.New().String()
		p.DefinitionID = definitionID
		p.FuelAPIID = null.StringFrom(img.FuelAPIID)
		p.Width = null.IntFrom(img.Width)
		p.Height = null.IntFrom(img.Height)
		p.SourceURL = device.SourceURL
		//p.DimoS3URL = null.StringFrom("") // dont set it so it is null
		p.Color = device.Color
		p.NotExactImage = img.NotExactImage

		err := p.Upsert(ctx, ch.dbs().Writer, true, []string{models.ImageColumns.DefinitionID, models.ImageColumns.SourceURL}, boil.Infer(), boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}

func getDefaultImage(img gateways.FuelDeviceImages) string {
	for _, image := range img.Images {
		return image.SourceURL
	}
	return ""
}
