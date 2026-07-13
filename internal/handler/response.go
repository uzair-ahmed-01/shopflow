package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// SendJSON sends a success envelope JSON response.
func SendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
	})
}

// SendError sends an error envelope JSON response and logs internal errors.
func SendError(w http.ResponseWriter, status int, message string, code string, errs ...error) {
	if len(errs) > 0 && errs[0] != nil {
		log.Error().Err(errs[0]).Str("code", code).Int("status", status).Msg(message)
	} else if status >= 500 {
		log.Error().Str("code", code).Int("status", status).Msg(message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]string{
			"message": message,
			"code":    code,
		},
	})
}

// DecodeJSON strictly decodes request body JSON into target generic type,
// limiting payload size to 1MB to protect against Denial of Service (DoS) attacks.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var val T
	// Limit body size to 1MB (1,048,576 bytes)
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Reject request if fields are unknown or typoed

	if err := dec.Decode(&val); err != nil {
		SendError(w, http.StatusBadRequest, "invalid request payload: "+err.Error(), "BAD_REQUEST")
		return val, false
	}
	return val, true
}

// SuccessResponse represents a generic success response envelope for Swagger.
type SuccessResponse[T any] struct {
	Success bool `json:"success" example:"true"`
	Data    T    `json:"data"`
}

// ErrorDetail represents the details of an error.
type ErrorDetail struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ErrorResponse represents a generic error response envelope.
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Error   ErrorDetail `json:"error"`
}

