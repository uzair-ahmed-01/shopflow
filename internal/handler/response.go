package handler

import (
	"encoding/json"
	"net/http"
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

// SendError sends an error envelope JSON response.
func SendError(w http.ResponseWriter, status int, message string, code string) {
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
