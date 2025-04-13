package validation

import (
	"github.com/go-playground/validator/v10"
)

var (
	Validator = validator.New()
)

func ValidateStruct(s interface{}) error {
	return Validator.Struct(s)
}
