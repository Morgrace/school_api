package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"simpleapi/internal/models"
)

// func ErrorHandler(err error, message string) error {
// 	errorLogger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
// 	errorLogger.Println(message, err)
// 	return fmt.Errorf(message)
// }

// ResponseError inspects the error to set the status code,
// but allows you to override the client-facing message.
func ResponseError(w http.ResponseWriter, err error, message string) {
	// 1. Check: Is it a 404 Not Found?
	if errors.Is(err, models.ErrNotFound) {
		if message == "" {
			message = err.Error() // Default
		}
		WriteError(w, http.StatusNotFound, message)
		return
	}

	// 2. Check: Is it a 409 Conflict?
	if errors.Is(err, models.ErrConflict) {
		if message == "" {
			message = err.Error() // Default
		}
		WriteError(w, http.StatusConflict, message)
		return
	}

	// 3. Check: Is it a 400 Bad Request? (e.g. Validation)
	if errors.Is(err, models.ErrInvalidInput) {
		if message == "" {
			message = err.Error() // Default
		}
		WriteError(w, http.StatusBadRequest, message)
		return
	}

	// For security, we usually ignore the raw error text for 500s.
	// If you passed a specific message (e.g. "Could not process upload"), we use it.
	// Otherwise, we default to "Internal Server Error".
	if message == "" {
		message = "Internal Server Error"
	}

	WriteError(w, http.StatusInternalServerError, message)
}

// WriteError sends the JSON response (The "Dumb" Formatter)
func WriteError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	status := "error"
	if code >= 400 && code < 500 {
		status = "fail"
	}

	json.NewEncoder(w).Encode(struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: message,
	})
}

// WriteJSON sends success response
func WriteJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
		Data   any    `json:"data"`
	}{
		Status: "success",
		Data:   data,
	})
}
