package common

import (
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/gofiber/fiber/v2"
)

func BindRequestPayload[S any](c *fiber.Ctx) *S {
	s := new(S)
	if err := c.BodyParser(s); err != nil {
		panic(&exceptions.ValidationError{
			Err: err,
		})
	}
	return s
}
