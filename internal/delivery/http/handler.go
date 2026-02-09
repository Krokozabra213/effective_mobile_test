// Package handler provides HTTP handlers for API.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Krokozabra213/effective_mobile/internal/domain"
	"github.com/google/uuid"
)

// Business defines business layer interface.
type Business interface {
	CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error)
	ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error)
	ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error)
	DeleteSubscription(ctx context.Context, id int64) error
	CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (domain.TotalCost, error)
}

// Handler handles HTTP requests.
type Handler struct {
	business Business
}

// New creates a new Handler and registers routes.
func New(mux *http.ServeMux, business Business) *Handler {
	h := &Handler{
		business: business,
	}

	// Subscriptions CRUD
	mux.HandleFunc("POST /subscriptions", h.CreateSubscription)
	mux.HandleFunc("GET /subscriptions/{id}", h.GetSubscriptionByID)
	mux.HandleFunc("GET /subscriptions", h.ListSubscriptions)
	mux.HandleFunc("PATCH /subscriptions/{id}", h.UpdateSubscription)
	mux.HandleFunc("DELETE /subscriptions/{id}", h.DeleteSubscription)
	mux.HandleFunc("GET /subscriptions/cost", h.CalculateTotalCost)

	// User subscriptions
	mux.HandleFunc("GET /users/{user_id}/subscriptions", h.ListSubscriptionsByUserID)

	return h
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidBody)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := parseMonthYear(*req.EndDate)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
			return
		}
		endDate = &parsed
	}

	input := domain.NewCreateSubscriptionInput(req.ServiceName, req.Price, req.UserID, startDate, endDate)

	sub, err := h.business.CreateSubscription(r.Context(), &input)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}
	resp := h.toSubscriptionResponse(sub)

	h.respondJSON(w, http.StatusCreated, &resp)
}

func (h *Handler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidIDFormat)
		return
	}

	if id <= 0 {
		h.respondError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	sub, err := h.business.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, h.toSubscriptionResponse(sub))
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	params := h.parsePagination(r)

	subs, err := h.business.ListSubscriptions(r.Context(), params)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	response := map[string]any{
		"subscriptions": h.toSubscriptionListResponse(subs),
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) ListSubscriptionsByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidUserIDFormat)
		return
	}

	params := h.parsePagination(r)

	subs, err := h.business.ListSubscriptionsByUserID(r.Context(), userID, params)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	response := map[string]any{
		"subscriptions": h.toSubscriptionListResponse(subs),
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidIDFormat)
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidBody)
		return
	}

	input := domain.UpdateSubscriptionInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
	}

	if req.EndDate != nil {
		parsed, err := parseMonthYear(*req.EndDate)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
			return
		}
		input.EndDate = &parsed
	}

	sub, err := h.business.UpdateSubscription(r.Context(), id, input)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, h.toSubscriptionResponse(sub))
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidIDFormat)
		return
	}

	if err := h.business.DeleteSubscription(r.Context(), id); err != nil {
		h.handleBusinessError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CalculateTotalCost(w http.ResponseWriter, r *http.Request) {
	startPeriod, err := parseMonthYear(r.URL.Query().Get("start_period"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
		return
	}

	endPeriod, err := parseMonthYear(r.URL.Query().Get("end_period"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
		return
	}

	filter := domain.CostFilter{
		StartPeriod: startPeriod,
		EndPeriod:   endPeriod,
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, ErrInvalidUserIDFormat)
			return
		}
		filter.UserID = &userID
	}

	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	result, err := h.business.CalculateTotalCost(r.Context(), filter)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, NewTotalCostResponse(result.TotalCost, result.Count))
}
