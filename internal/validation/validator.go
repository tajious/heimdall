package validation

import (
	"github.com/go-playground/validator/v10"
)

var (
	// Validator is a shared validator instance
	Validator = validator.New()
)

// ValidateStruct validates a struct using the shared validator
func ValidateStruct(s interface{}) error {
	return Validator.Struct(s)
}
