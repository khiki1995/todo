package helper

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

var validate = validator.New()

func ValidateStruct(obj interface{}) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(obj)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

func ResponseData(ctx *fiber.Ctx, data interface{}) error {
	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"success": true,
		"data": data,
	})
}

func GetMax(v1, v2 int) int {
	if v1 > v2 {
		return v1
	}
	return v2
}

func GetMin(v1, v2 int) int {
	if v1 > v2 {
		return v2
	}
	return v1
}