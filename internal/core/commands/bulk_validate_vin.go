//nolint:tagliatelle
package commands

import (
	"context"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
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
	VIN          string                  `json:"vin"`
	DefinitionID string                  `json:"definition_id"`
	DeviceMake   coremodels.Manufacturer `json:"device_make"`
	DeviceYear   int32                   `json:"device_year"`
	DeviceModel  string                  `json:"device_model"`
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

		devideDefinition, err := dc.DeviceDefinitionDataHandler.Handle(ctx, &queries.GetDeviceDefinitionByIDQuery{DeviceDefinitionID: decodedVIN.DefinitionId}) //nolint

		if err == nil {
			row, rowErr := decodedVINFrom(vin, decodedVIN, devideDefinition)
			if rowErr != nil {
				return nil, rowErr
			}
			decodedVINs = append(decodedVINs, row)
		}
	}

	response := BulkValidateVinCommandResult{
		DecodedVINs:    decodedVINs,
		NotDecodedVins: notDecodedVins,
	}

	return response, nil
}

// decodedVINFrom builds one result row out of the decode and whatever
// GetDeviceDefinitionByIDQuery answered for its definition id.
//
// That handler answers the catalog's *coremodels.Template. The checked form
// is deliberate: an unchecked assertion here took the whole request down with
// a panic on the SUCCESS path -- every VIN that decoded and whose definition
// was in the catalog -- because it named a result type the handler had
// stopped returning. A surprise type is a broken contract worth reporting,
// not worth crashing over.
//
// DeviceModel is the template's model. It used to be the first device style's
// sub-model, which was both a different thing from the field's name and an
// unchecked index into a slice that is empty for any definition with no
// styles.
func decodedVINFrom(vin string, decoded *p_grpc.DecodeVinResponse, definition interface{}) (DecodedVIN, error) {
	tmpl, ok := definition.(*coremodels.Template)
	if !ok || tmpl == nil {
		return DecodedVIN{}, fmt.Errorf("device definition %s: expected *models.Template from GetDeviceDefinitionByIDQuery, got %T", decoded.DefinitionId, definition)
	}
	return DecodedVIN{
		VIN:          vin,
		DefinitionID: decoded.DefinitionId,
		DeviceYear:   decoded.Year,
		DeviceMake: coremodels.Manufacturer{
			TokenID: tmpl.Manufacturer.TokenID,
			Name:    tmpl.Manufacturer.Name,
		},
		DeviceModel: tmpl.Model,
	}, nil
}
