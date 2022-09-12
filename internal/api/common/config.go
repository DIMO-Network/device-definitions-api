package common

import (
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FiberConfig(disableStartupMsg bool) fiber.Config {
	return fiber.Config{
		DisableStartupMessage: disableStartupMsg,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {

			code := fiber.StatusInternalServerError
			_type := "https://tools.ietf.org/html/rfc7231#section-6.6.1"
			title := "An error occurred while processing your request."

			if _, ok := err.(*exceptions.ValidationError); ok {
				code = fiber.StatusBadRequest
				_type = "https://tools.ietf.org/html/rfc7231#section-6.5.1"
				title = "The specified resource is not valid."
			}

			if _, ok := err.(*exceptions.NotFoundError); ok {
				code = fiber.StatusNotFound
				_type = "https://tools.ietf.org/html/rfc7231#section-6.5.4"
				title = "The specified resource was not found."
			}

			if _, ok := err.(*exceptions.ConflictError); ok {
				code = fiber.StatusConflict
				_type = "https://tools.ietf.org/html/rfc7231#section-6.5.1"
				title = "The specified resource is not valid."
			}

			p := &ProblemDetails{
				Type:   _type,
				Title:  title,
				Status: code,
				Detail: err.Error(),
			}

			return ctx.Status(code).JSON(p)
		},
	}
}

func GrpcConfig(p any) (err error) {

	if e, ok := p.(*exceptions.ValidationError); ok {
		return status.Errorf(codes.InvalidArgument, e.Error())
	}

	if e, ok := p.(*exceptions.NotFoundError); ok {
		return status.Errorf(codes.NotFound, e.Error())
	}

	if e, ok := p.(*exceptions.ConflictError); ok {
		return status.Errorf(codes.Aborted, e.Error())
	}

	return status.Errorf(codes.Internal, "An error occurred while processing your request: %v", p)
}
