package common

import (
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/go-playground/validator"
)

func Validate(s interface{}) []*common.ErrorResponse {
	var validate = validator.New()
	var errors []*common.ErrorResponse
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element common.ErrorResponse
			element.Field = err.Field()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
