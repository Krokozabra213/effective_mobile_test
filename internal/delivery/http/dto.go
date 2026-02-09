package handler

import (
	"errors"

	"github.com/google/uuid"
)

var ErrPrice = errors.New("price should be >=0")

// ===== Request DTOs =====

// CreateSubscriptionRequest запрос на создание подписки
type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int32     `json:"price" example:"400"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
}

func (r CreateSubscriptionRequest) Validate() error {
	if r.Price < 0 {
		return ErrPrice
	}
	return nil
}

// UpdateSubscriptionRequest запрос на обновление подписки
type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Netflix"`
	Price       *int    `json:"price,omitempty" example:"800"`
	EndDate     *string `json:"end_date,omitempty" example:"12-2025"`
}

func (r UpdateSubscriptionRequest) Validate() error {
	if *r.Price < 0 {
		return ErrPrice
	}
	return nil
}

// ===== Response DTOs =====

type SubscriptionResponse struct {
	ID          int64     `json:"id" example:"1"`
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int32     `json:"price" example:"400"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     *string   `json:"end_date,omitempty" example:"12-2025"`
	CreatedAt   string    `json:"created_at" example:"2025-01-15T10:30:00Z"`
}

// TotalCostResponse ответ с суммарной стоимостью
type TotalCostResponse struct {
	TotalCost int64 `json:"total_cost" example:"1200"`
	Count     int64 `json:"count" example:"3"`
}

func NewTotalCostResponse(totalCost int64, count int64) TotalCostResponse {
	return TotalCostResponse{
		TotalCost: totalCost,
		Count:     count,
	}
}
