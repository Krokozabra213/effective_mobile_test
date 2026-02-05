package domain

import (
	"time"

	"github.com/google/uuid"
)

// ===== Request DTOs =====

// CreateSubscriptionRequest запрос на создание подписки
type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int32     `json:"price" example:"400"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
}

type CalculateCostRequest struct {
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
}

// ===== Response DTOs =====

type SubscriptionResponse struct {
	ID          int64     `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}
