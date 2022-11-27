package queries

import (
	"context"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	p_grpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
)

type DecodeVINQueryHandler struct {
	DBS func() *db.ReaderWriter
}

type DecodeVINQuery struct {
	VIN string `json:"vin"`
}

func (*DecodeVINQuery) Key() string { return "DecodeVINQuery" }

func NewDecodeVINQueryHandler(dbs func() *db.ReaderWriter) DecodeVINQueryHandler {
	return DecodeVINQueryHandler{
		DBS: dbs,
	}
}

func (dc DecodeVINQueryHandler) Handle(ctx context.Context, query mediator.Message) (interface{}, error) {
	qry := query.(*DecodeVINQuery)
	if len(qry.VIN) != 17 {
		return nil, &exceptions.ValidationError{Err: fmt.Errorf("invalid vin %s", qry.VIN)}
	}
	resp := p_grpc.DecodeVINResponse{}
	// todo implement, write a test for this once have DB structure
	//
	wmi := qry.VIN[0:3]
	// query device makes where find wmi - what's best datastructure to query for this? change to array varchar(3)
	resp.DeviceMakeId = wmi // todo
	vin := shared.VIN(qry.VIN)
	resp.Year = int32(vin.Year()) // needs to be updated for newer years
	// lookup the device definition by rest of info - look at our existing vins

	// if no match, call drivly, and then update our database (wmi if no match, model)
	// make drivly a service

	return resp, nil
}
