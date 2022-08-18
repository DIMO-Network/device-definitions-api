package common

import (
	"context"

	"github.com/TheFellow/go-mediator/mediator"
	"github.com/gofiber/fiber/v2"
)

type LoggingBehavior struct{}

func (p LoggingBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {

	// _, span := trace.NewSpan(ctx, fmt.Sprintf("dimo device-definitions-api request : %v - %+v", msg.Key(), msg))
	// defer span.End()

	return next(ctx)
}

type ValidationBehavior struct{}

func (p ValidationBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {

	valErrors := Validate(msg)
	if valErrors != nil {
		panic(fiber.NewError(400, valErrors[0].Field))
	}
	return next(ctx)
}

type ErrorHandlingBehavior struct{}

func (p ErrorHandlingBehavior) Process(ctx context.Context, msg mediator.Message, next mediator.Next) (interface{}, error) {

	r, err := next(ctx)
	if err != nil {
		panic(err)
	}

	return r, nil
}
