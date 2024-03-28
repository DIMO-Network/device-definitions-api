//nolint:tagliatelle
package commands

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/db"
)

type BulkValidateVinCommand struct {
	VINs []string `json:"vins"`
}

type BulkValidateVinCommandResult struct {
	DecodedVINs    []DecodedVIN `json:"decoded_vins"`
	NotDecodedVins []string     `json:"not_decoded_vins"`
}

type DecodedVIN struct {
	VIN                   string                          `json:"vin"`
	DeviceDefinitionID    string                          `json:"device_definition_id"`
	DeviceMake            models.DeviceMake               `json:"device_make"`
	DeviceYear            int32                           `json:"device_year"`
	DeviceModel           string                          `json:"device_model"`
	CompatibilityFeatures []*p_grpc.DeviceCompatibilities `json:"compatibility_features"`
}

func (*BulkValidateVinCommand) Key() string { return "BulkValidateVinCommand" }

type BulkValidateVinCommandHandler struct {
	DBS                                  func() *db.ReaderWriter
	DecodeVINHandler                     queries.DecodeVINQueryHandler
	DeviceDefinitionCompatibilityHandler queries.GetCompatibilityByDeviceDefinitionQueryHandler
	DeviceDefinitionDataHandler          queries.GetDeviceDefinitionByIDQueryHandler
}

func NewBulkValidateVinCommandHandler(dbs func() *db.ReaderWriter, decodeVINHandler queries.DecodeVINQueryHandler, deviceDefinitionCompatibilityHandler queries.GetCompatibilityByDeviceDefinitionQueryHandler, deviceDefintionDataHandler queries.GetDeviceDefinitionByIDQueryHandler) BulkValidateVinCommandHandler {
	return BulkValidateVinCommandHandler{
		DBS:                                  dbs,
		DecodeVINHandler:                     decodeVINHandler,
		DeviceDefinitionCompatibilityHandler: deviceDefinitionCompatibilityHandler,
		DeviceDefinitionDataHandler:          deviceDefintionDataHandler,
	}
}

func (dc BulkValidateVinCommandHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	command := query.(*BulkValidateVinCommand)

	if len(command.VINs) == 0 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("cannot decode vin array of %s", command.VINs)}
	}

	decodedVINs := make([]DecodedVIN, 0)
	notDecodedVins := make([]string, 0)

	for _, vin := range command.VINs {
		decodedVIN, err := dc.DecodeVINHandler.Handle(ctx, &queries.DecodeVINQuery{VIN: vin})
		if err != nil {
			notDecodedVins = append(notDecodedVins, vin)
			continue
		}

		compatibilities := make([]*p_grpc.DeviceCompatibilities, 0)

		deviceDefinitionCompatibilities, err := dc.DeviceDefinitionCompatibilityHandler.Handle(ctx, &queries.GetCompatibilityByDeviceDefinitionQuery{DeviceDefinitionID: decodedVIN.(*p_grpc.DecodeVinResponse).DeviceDefinitionId})

		if err == nil {
			compatibilities = deviceDefinitionCompatibilities.(*p_grpc.GetDeviceCompatibilitiesResponse).Compatibilities
		}

		devideDefinition, err := dc.DeviceDefinitionDataHandler.Handle(ctx, &queries.GetDeviceDefinitionByIDQuery{DeviceDefinitionID: decodedVIN.(*p_grpc.DecodeVinResponse).DeviceDefinitionId})

		if err == nil {
			decodedVINs = append(decodedVINs, DecodedVIN{
				VIN:                   vin,
				DeviceDefinitionID:    decodedVIN.(*p_grpc.DecodeVinResponse).DeviceDefinitionId,
				DeviceYear:            decodedVIN.(*p_grpc.DecodeVinResponse).Year,
				CompatibilityFeatures: compatibilities,
				DeviceMake:            devideDefinition.(*models.GetDeviceDefinitionQueryResult).DeviceMake,
				DeviceModel:           devideDefinition.(*models.GetDeviceDefinitionQueryResult).DeviceStyles[0].SubModel,
			})
		}
	}

	response := BulkValidateVinCommandResult{
		DecodedVINs:    decodedVINs,
		NotDecodedVins: notDecodedVins,
	}

	return response, nil
}
