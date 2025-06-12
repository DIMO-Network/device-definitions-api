//nolint:tagliatelle
package commands

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	pgrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
)

type BulkValidateVinCommand struct {
	VINs []string `json:"vins"`
}

type BulkValidateVinCommandResult struct {
	DecodedVINs    []DecodedVIN `json:"decoded_vins"`
	NotDecodedVins []string     `json:"not_decoded_vins"`
}

type DecodedVIN struct {
	VIN          string                `json:"vin"`
	DefinitionID string                `json:"definition_id"`
	DeviceMake   gateways.Manufacturer `json:"device_make"`
	DeviceYear   int32                 `json:"device_year"`
	DeviceModel  string                `json:"device_model"`
}

func (*BulkValidateVinCommand) Key() string { return "BulkValidateVinCommand" }

type BulkValidateVinCommandHandler struct {
	DBS                         func() *db.ReaderWriter
	DecodeVINHandler            queries.DecodeVINQueryHandler
	DeviceDefinitionDataHandler queries.GetDeviceDefinitionByIDQueryHandler
}

func NewBulkValidateVinCommandHandler(dbs func() *db.ReaderWriter, decodeVINHandler queries.DecodeVINQueryHandler, deviceDefintionDataHandler queries.GetDeviceDefinitionByIDQueryHandler) BulkValidateVinCommandHandler {
	return BulkValidateVinCommandHandler{
		DBS:                         dbs,
		DecodeVINHandler:            decodeVINHandler,
		DeviceDefinitionDataHandler: deviceDefintionDataHandler,
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

		devideDefinition, err := dc.DeviceDefinitionDataHandler.Handle(ctx, &queries.GetDeviceDefinitionByIDQuery{DeviceDefinitionID: decodedVIN.(*pgrpc.DecodeVinResponse).DefinitionId}) //nolint

		if err == nil {
			dd := devideDefinition.(*models.GetDeviceDefinitionQueryResult)
			dm := gateways.Manufacturer{
				TokenID: dd.MakeTokenID,
				Name:    dd.MakeName,
			}

			decodedVINs = append(decodedVINs, DecodedVIN{
				VIN:          vin,
				DefinitionID: decodedVIN.(*pgrpc.DecodeVinResponse).DefinitionId,
				DeviceYear:   decodedVIN.(*pgrpc.DecodeVinResponse).Year,
				DeviceMake:   dm,
				DeviceModel:  devideDefinition.(*models.GetDeviceDefinitionQueryResult).DeviceStyles[0].SubModel,
			})
		}
	}

	response := BulkValidateVinCommandResult{
		DecodedVINs:    decodedVINs,
		NotDecodedVins: notDecodedVins,
	}

	return response, nil
}
