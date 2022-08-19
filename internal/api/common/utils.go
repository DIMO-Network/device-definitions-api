package common

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/gofiber/fiber/v2"
)

func BindRequestPayload[S any](c *fiber.Ctx) *S {
	s := new(S)
	if err := c.BodyParser(s); err != nil {
		panic(&common.ValidationError{
			Err: err,
		})
	}
	return s
}
