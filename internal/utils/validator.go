package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ValidationError map[string]string

func ValidateStruct(data interface{}) ValidationError {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	errors := ValidationError{}
	for _, e := range err.(validator.ValidationErrors) {

		var msg string

		switch e.Tag() {
		case "required":
			msg = "field is required"
		case "email":
			msg = "invalid email format"
		case "min":
			msg = fmt.Sprintf("minimum %s characters", e.Param())
		case "max":
			msg = fmt.Sprintf("maximum %s characters", e.Param())
		default:
			msg = "invalid value"
		}

		errors[e.Field()] = msg
	}

	return errors
}


func ValidateURL(u string) bool {
    parsed, err := url.ParseRequestURI(u)
    if err != nil {
        return false
    }

    if !strings.HasPrefix(parsed.Scheme, "http") {
        return false
    }

    return true
}