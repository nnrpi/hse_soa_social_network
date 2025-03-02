package api

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	errorMessages := []string{}

	for _, e := range validationErrors {
		switch e.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required", e.Field()))
		case "email":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid email address", e.Field()))
		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be min %s characters long", e.Field(), e.Param()))
		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must not max %s characters long", e.Field(), e.Param()))
		case "datetime":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid date in format %s", e.Field(), e.Param()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation: %s", e.Field(), e.Tag()))
		}
	}

	return fmt.Errorf(strings.Join(errorMessages, "; "))
}
