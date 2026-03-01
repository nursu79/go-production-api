package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nursu79/go-production-api/internal/domain"
)

// RespondJSON sends a JSON response with the given status code.
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to marshal response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// RespondError maps domain errors to standard HTTP status codes.
func RespondError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Internal server error"

	switch {
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
		message = "Invalid input"
	case errors.Is(err, domain.ErrDuplicateEmail):
		status = http.StatusConflict
		message = "Email already registered"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = "Resource not found"
	}

	RespondJSON(w, status, map[string]string{"error": message})
}
