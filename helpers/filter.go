package helpers

import (
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func Filter(model interface{}, fields []string) (interface{}, error) {
	requiredFields := make(map[string]bool)
	for _, field := range fields {
		requiredFields[field] = true
	}

	value := reflect.ValueOf(model)
	if value.Kind() != reflect.Ptr || value.IsNil() {

		return nil, &fiber.Error{Code: 500, Message: "Model must be a non-nil pointer."}
	}

	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return nil, &fiber.Error{Code: 500, Message: "Model must be a struct."}
	}

	filteredModel := reflect.New(value.Type()).Interface()

	filteredValue := reflect.ValueOf(filteredModel).Elem()
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		fieldName := field.Name

		if requiredFields[fieldName] {
			fieldValue := value.Field(i)
			filteredValue.FieldByName(fieldName).Set(fieldValue)
		}
	}

	return filteredModel, nil
}
