package handlers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var requestValidator = validator.New()

func parseAndValidateJSON(c *fiber.Ctx, payload interface{}) error {
	if err := c.BodyParser(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := requestValidator.Struct(payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, humanizeValidationError(err))
	}
	return nil
}

func humanizeValidationError(err error) string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		fieldErr := validationErrs[0]
		field := fieldErr.Field()
		return fmt.Sprintf("%s failed on the '%s' rule", toSnakeCase(field), fieldErr.Tag())
	}
	return err.Error()
}

func toSnakeCase(input string) string {
	var builder strings.Builder
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteByte('_')
		}
		builder.WriteRune(r)
	}
	return strings.ToLower(builder.String())
}

