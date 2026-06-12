package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/taskmanager/backend/internal/models"
)

// writeJSON sends a JSON response with the given status code and payload.
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// writeInternalError sends a generic 500 Internal Server Error response.
func writeInternalError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusInternalServerError, models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// writeUnauthorizedError sends a 401 Unauthorized response.
func writeUnauthorizedError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// writeNotFoundError sends a 404 Not Found response.
func writeNotFoundError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusNotFound, models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// writeValidationError sends a 422 Unprocessable Entity response for a single field error.
func writeValidationError(w http.ResponseWriter, message string, detail models.ErrorDetail) {
	writeJSON(w, http.StatusUnprocessableEntity, models.ValidationErrorResponse{
		Success: false,
		Message: message,
		Errors:  []models.ErrorDetail{detail},
	})
}

// writeValidationErrors sends a 422 Unprocessable Entity response for multiple field errors.
func writeValidationErrors(w http.ResponseWriter, err error) {
	var errorsList []models.ErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, ve := range validationErrors {
			errorsList = append(errorsList, models.ErrorDetail{
				Field:   ve.Field(),
				Message: formatValidationError(ve),
			})
		}
	}

	writeJSON(w, http.StatusUnprocessableEntity, models.ValidationErrorResponse{
		Success: false,
		Message: "validation failed",
		Errors:  errorsList,
	})
}

// formatValidationError converts a validator field error into a human-readable message.
func formatValidationError(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + ve.Param() + " characters"
	case "max":
		return "must not exceed " + ve.Param() + " characters"
	case "oneof":
		return "must be one of: " + ve.Param()
	case "uuid":
		return "must be a valid UUID"
	case "gte":
		return "must be greater than or equal to " + ve.Param()
	case "lte":
		return "must be less than or equal to " + ve.Param()
	default:
		return "validation failed on " + ve.Tag()
	}
}