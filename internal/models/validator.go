package models

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError is your clean, public-facing error format
type ValidationError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
	Index *int   `json:"index,omitempty"` // Pointer so it's null if not applicable
}

// 1. Helper: Converts raw validator engine errors into your clean format
func parseValidationErr(err error, index *int) []ValidationError {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		out := make([]ValidationError, len(validationErrors))
		for i, fieldErr := range validationErrors {
			out[i] = ValidationError{
				Field: fieldErr.Field(),
				Msg:   msgForTag(fieldErr.Tag()),
				Index: index, // Adds index only if it exists
			}
		}
		return out
	}
	return nil
}

// 2. ValidateOne: Validates a single struct
func ValidateOne(s any) []ValidationError {
	err := validate.Struct(s)
	if err != nil {
		return parseValidationErr(err, nil)
	}
	return nil
}

// 3. ValidateBatch: Validates a list of structs (Generic)
func ValidateBatch[T any](dataset []T) []ValidationError {
	var errorList []ValidationError

	for i, item := range dataset {
		err := validate.Struct(item)
		if err != nil {
			// pass the index to the parser
			idx := i
			batchErrors := parseValidationErr(err, &idx)
			if batchErrors != nil {
				errorList = append(errorList, batchErrors...)
			} else {
				log.Printf("Internal Validator Error at index %d: %v", i, err)
			}
		}
	}
	return errorList
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	}
	return "Invalid field"
}
