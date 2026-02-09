package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int64
	ServiceName string
	Price       int32
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
}

func NewSubscription(id int64, serviceName string, price int32, userID uuid.UUID, start time.Time,
	end *time.Time, createdAt time.Time,
) *Subscription {
	return &Subscription{
		ID:          id,
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
		CreatedAt:   createdAt,
	}
}

type CreateSubscriptionInput struct {
	ServiceName string
	Price       int32
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

func NewCreateSubscriptionInput(service string, price int32, userID uuid.UUID, start time.Time, end *time.Time) CreateSubscriptionInput {
	return CreateSubscriptionInput{
		ServiceName: service,
		Price:       price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}
}

type UpdateSubscriptionInput struct {
	ServiceName *string
	Price       *int
	EndDate     *time.Time
}

type CostFilter struct {
	StartPeriod time.Time // "01-2025"
	EndPeriod   time.Time // "12-2025"
	UserID      *uuid.UUID
	ServiceName *string
}

type TotalCost struct {
	TotalCost int64
	Count     int64
}

type ListParams struct {
	Limit  int32
	Offset int32
}
