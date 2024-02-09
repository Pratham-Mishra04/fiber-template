package helpers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func Validate[T any](payload T) error {
	validate := validator.New()

	if err := validate.Struct(payload); err != nil {
		validationErrors := err.(validator.ValidationErrors)

		var errorsBuilder strings.Builder

		for _, fieldError := range validationErrors {
			field := fieldError.Field()
			errorMessage := fmt.Sprintf("Invalid %s \n", field)
			errorsBuilder.WriteString(errorMessage)
		}

		errorsString := errorsBuilder.String()

		return &fiber.Error{Code: 400, Message: errorsString}
	}
	return nil
}
