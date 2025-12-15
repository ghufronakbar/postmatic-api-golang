// pkg/utils/validator.go

package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Inisialisasi validator instance (singleton)
var validate = validator.New()

type ErrorResponse map[string]string

// ValidateStruct memvalidasi struct dan me-return map error jika ada
func ValidateStruct(s interface{}) ErrorResponse {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(ErrorResponse)
	for _, err := range err.(validator.ValidationErrors) {
		// Custom message logic bisa ditaruh disini
		field := strings.ToLower(err.Field())

		// Contoh formatting pesan error
		switch err.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errors[field] = "Invalid email format"
		case "gte":
			errors[field] = fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
		default:
			errors[field] = fmt.Sprintf("Invalid value for %s", field)
		}
	}
	return errors
}
