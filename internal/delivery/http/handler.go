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
	mux.HandleFunc("POST /subscriptions/cost", h.CalculateTotalCost)

	// User subscriptions
	mux.HandleFunc("GET /users/{user_id}/subscriptions", h.ListSubscriptionsByUserID)

	return h
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidBody)
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

    input:=domain.NewCreateSubscriptionInput(req.ServiceName, req.Price, req.UserID, startDate, endDate)

	sub, err := h.business.CreateSubscription(r.Context(), &input)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toSubscriptionResponse(sub))
}
