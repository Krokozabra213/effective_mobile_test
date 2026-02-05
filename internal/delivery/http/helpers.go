package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error" example:"subscription not found"`
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, ErrorResponse{Error: message})
}
