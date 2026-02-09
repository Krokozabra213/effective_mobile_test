package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Krokozabra213/effective_mobile/internal/business"
	"github.com/Krokozabra213/effective_mobile/internal/domain"
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

func (h *Handler) toSubscriptionResponse(sub *domain.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   formatMonthYear(sub.StartDate),
		CreatedAt:   sub.CreatedAt.Format(time.RFC3339),
	}

	if sub.EndDate != nil {
		formatted := formatMonthYear(*sub.EndDate)
		resp.EndDate = &formatted
	}

	return resp
}

func (h *Handler) parseID(r *http.Request, param string) (int64, error) {
	idStr := r.PathValue(param)
	return strconv.ParseInt(idStr, 10, 64)
}

func (h *Handler) parsePagination(r *http.Request) domain.ListParams {
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return domain.ListParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}

func (h *Handler) toSubscriptionListResponse(subs []domain.Subscription) []SubscriptionResponse {
	result := make([]SubscriptionResponse, len(subs))
	for i, sub := range subs {
		result[i] = h.toSubscriptionResponse(&sub)
	}
	return result
}

func (h *Handler) handleBusinessError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, business.ErrNotFound):
		h.respondError(w, http.StatusNotFound, "subscription not found")
	default:
		h.respondError(w, http.StatusInternalServerError, "internal error")
	}
}
