package fortify

import (
	"encoding/json"
	"net/http"
)

type jsonMessage struct {
	OK      bool                `json:"ok,omitempty"`
	Message string              `json:"message,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

func writeOK(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusOK, jsonMessage{OK: true, Message: message})
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func writeValidation(w http.ResponseWriter, field, message string) {
	writeJSON(w, http.StatusUnprocessableEntity, jsonMessage{
		Message: "validation failed",
		Errors:  map[string][]string{field: []string{message}},
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, jsonMessage{Message: message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
