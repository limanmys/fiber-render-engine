package validator

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validation for base64
	validate.RegisterValidation("base64", validateBase64)

	// Use json tag names for validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func validateBase64(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(value)
	return err == nil
}

func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				errors = append(errors, fmt.Sprintf("%s is required", err.Field()))
			case "base64":
				errors = append(errors, fmt.Sprintf("%s must be valid base64", err.Field()))
			default:
				errors = append(errors, fmt.Sprintf("%s failed validation %s", err.Field(), err.Tag()))
			}
		}
		return fmt.Errorf("%s", strings.Join(errors, ", "))
	}
	return nil
}
