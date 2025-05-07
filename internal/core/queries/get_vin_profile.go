//nolint:tagliatelle
package queries

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type GetVINProfileQueryHandler struct {
	dbs    func() *db.ReaderWriter
	logger *zerolog.Logger
}

type GetVINProfileQuery struct {
	VIN string `json:"vin"`
}

type GetVINProfileResponse struct {
	VIN        string `json:"vin"`
	ProfileRaw []byte `json:"profileRaw"`
	Vendor     string `json:"vendor"`
}

func (*GetVINProfileQuery) Key() string { return "GetVINProfileQuery" }

func NewGetVINProfileQueryHandler(dbs func() *db.ReaderWriter, logger *zerolog.Logger) GetVINProfileQueryHandler {
	return GetVINProfileQueryHandler{
		dbs:    dbs,
		logger: logger,
	}
}

func (dc GetVINProfileQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*GetVINProfileQuery)
	if len(qry.VIN) < 13 || len(qry.VIN) > 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}

	vinNumber, err := models.VinNumbers(models.VinNumberWhere.Vin.EQ(qry.VIN), models.VinNumberWhere.DrivlyData.IsNotNull()).One(ctx, dc.dbs().Reader)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &exceptions.NotFoundError{Err: fmt.Errorf("vin %s not found", qry.VIN)}
		}
		return nil, &exceptions.InternalError{Err: fmt.Errorf("failed to get vin %s", qry.VIN)}
	}

	return &GetVINProfileResponse{
		VIN:        qry.VIN,
		ProfileRaw: vinNumber.DrivlyData.JSON,
		Vendor:     "drivly",
	}, nil
}
