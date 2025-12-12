package server

import "strings"

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// validateUserInput перевіряє поля name, email, age
func validateUserInput(name, email string, age int32) []ValidationError {
	var errs []ValidationError

	if strings.TrimSpace(name) == "" {
		errs = append(errs, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(name) > 255 {
		errs = append(errs, ValidationError{
			Field:   "name",
			Message: "name is too long (max 255)",
		})
	}

	if strings.TrimSpace(email) == "" {
		errs = append(errs, ValidationError{
			Field:   "email",
			Message: "email is required",
		})
	} else if len(email) > 255 {
		errs = append(errs, ValidationError{
			Field:   "email",
			Message: "email is too long (max 255)",
		})
	}

	if age < 18 {
		errs = append(errs, ValidationError{
			Field:   "age",
			Message: "age must be >= 18",
		})
	}

	return errs
}
