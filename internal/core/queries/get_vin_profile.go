//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/shared/pkg/logfields"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetVINProfileQueryHandler struct {
	dbs           func() *db.ReaderWriter
	logger        *zerolog.Logger
	powertrainSvc services.PowerTrainTypeService
}

type GetVINProfileQuery struct {
	VIN string `json:"vin"`
}

type GetVINProfileResponse struct {
	VIN            string `json:"vin"`
	ProfileRaw     []byte `json:"profileRaw"`
	Vendor         string `json:"vendor"`
	PowertrainType string `json:"powertrainType,omitempty"`
}

func (*GetVINProfileQuery) Key() string { return "GetVINProfileQuery" }

func NewGetVINProfileQueryHandler(dbs func() *db.ReaderWriter, logger *zerolog.Logger, powertrainSvc services.PowerTrainTypeService) GetVINProfileQueryHandler {
	return GetVINProfileQueryHandler{
		dbs:           dbs,
		logger:        logger,
		powertrainSvc: powertrainSvc,
	}
}

func (dc GetVINProfileQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetVINProfileQuery)
	if len(qry.VIN) < 10 || len(qry.VIN) > 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}

	vinNumber, err := models.VinNumbers(models.VinNumberWhere.Vin.EQ(qry.VIN)).One(ctx, dc.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{Err: fmt.Errorf("vin %s not found", qry.VIN)}
		}
		return nil, &exceptions.InternalError{Err: fmt.Errorf("failed to get vin %s", qry.VIN)}
	}
	// try getting powertrain from override column in db
	powertrain := vinNumber.PowertrainType.String
	if powertrain == "" {
		// otherwise resolve from powertrain type
		split := strings.Split(vinNumber.DefinitionID, "_")
		if len(split) == 3 {
			powertrain, err = dc.powertrainSvc.ResolvePowerTrainType(split[0], split[1], vinNumber.DrivlyData, vinNumber.VincarioData)
			if err != nil {
				dc.logger.Error().Err(err).Str(logfields.VIN, qry.VIN).Msg("failed to resolve powertrain type for vin, continuing with ICE")
			}
		}
	}
	if powertrain == "" {
		powertrain = string(coremodels.ICE) //default to ICE
	}

	return &GetVINProfileResponse{
		VIN:            qry.VIN,
		ProfileRaw:     vinNumber.DrivlyData.JSON,
		Vendor:         vinNumber.DecodeProvider.String,
		PowertrainType: powertrain,
	}, nil
}
