package validation

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"agent/internal/repositories"
	"agent/internal/schemas"
)

var validate = validator.New()

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidatePayload validates a proposal payload against the appropriate schema
func ValidatePayload(entityType, operation string, payload []byte) []ValidationError {
	if operation == repositories.OperationDelete {
		return nil
	}

	target := getSchemaForEntity(entityType, operation)
	if target == nil {
		return nil
	}

	if err := json.Unmarshal(payload, target); err != nil {
		return []ValidationError{{
			Field:   "_payload",
			Tag:     "json",
			Message: fmt.Sprintf("invalid JSON: %s", err.Error()),
		}}
	}

	err := validate.Struct(target)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []ValidationError{{
			Field:   "_unknown",
			Tag:     "validation",
			Message: err.Error(),
		}}
	}

	return toValidationErrors(validationErrs)
}

func getSchemaForEntity(entityType, operation string) interface{} {
	if operation == repositories.OperationCreate {
		return getCreateSchema(entityType)
	}
	if operation == repositories.OperationUpdate {
		return getUpdateSchema(entityType)
	}
	return nil
}

func getCreateSchema(entityType string) interface{} {
	if entityType == "contact" {
		return &schemas.ContactCreate{}
	}
	if entityType == "organization" {
		return &schemas.OrganizationCreate{}
	}
	if entityType == "individual" {
		return &schemas.IndividualCreate{}
	}
	if entityType == "note" {
		return &schemas.NoteCreate{}
	}
	if entityType == "task" {
		return &schemas.TaskCreate{}
	}
	return nil
}

func getUpdateSchema(entityType string) interface{} {
	if entityType == "contact" {
		return &schemas.ContactUpdate{}
	}
	if entityType == "organization" {
		return &schemas.OrganizationUpdate{}
	}
	if entityType == "individual" {
		return &schemas.IndividualUpdate{}
	}
	if entityType == "note" {
		return &schemas.NoteUpdate{}
	}
	if entityType == "task" {
		return &schemas.TaskUpdate{}
	}
	return nil
}

func toValidationErrors(errs validator.ValidationErrors) []ValidationError {
	result := make([]ValidationError, len(errs))
	for i, err := range errs {
		result[i] = ValidationError{
			Field:   toSnakeCase(err.Field()),
			Tag:     err.Tag(),
			Message: buildErrorMessage(err),
		}
	}
	return result
}

func buildErrorMessage(err validator.FieldError) string {
	tag := err.Tag()
	if tag == "required" {
		return fmt.Sprintf("%s is required", toSnakeCase(err.Field()))
	}
	if tag == "email" {
		return "must be a valid email address"
	}
	if tag == "url" {
		return "must be a valid URL"
	}
	if tag == "e164" {
		return "must be a valid phone number in E.164 format"
	}
	if tag == "oneof" {
		return fmt.Sprintf("must be one of: %s", err.Param())
	}
	if tag == "min" {
		return fmt.Sprintf("must be at least %s", err.Param())
	}
	if tag == "max" {
		return fmt.Sprintf("must be at most %s", err.Param())
	}
	return fmt.Sprintf("failed %s validation", tag)
}

func toSnakeCase(s string) string {
	var result []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(c+'a'-'A'))
			continue
		}
		result = append(result, byte(c))
	}
	return string(result)
}
