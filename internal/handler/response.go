package handler

import (
	"encoding/json"
	"net/http"
)

// sendJSON sends a success envelope JSON response.
func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
	})
}

// sendError sends an error envelope JSON response.
func sendError(w http.ResponseWriter, status int, message string, code string) {
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
