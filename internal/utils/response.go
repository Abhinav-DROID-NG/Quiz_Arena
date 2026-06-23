package utils

import (
	"encoding/json"
	"net/http"
)

// WriteJSON serializes v as JSON and writes it with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error response with the given status and message.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// WriteSuccess writes a JSON success response with an optional message.
func WriteSuccess(w http.ResponseWriter, status int, msg string, data any) {
	payload := map[string]any{}
	if msg != "" {
		payload["message"] = msg
	}
	if data != nil {
		payload["data"] = data
	}
	WriteJSON(w, status, payload)
}
