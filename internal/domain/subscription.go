package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int64
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
}

type CreateSubscriptionInput struct {
    ServiceName string
    Price       int
    UserID      uuid.UUID
    StartDate   time.Time
    EndDate     *time.Time
}

type UpdateSubscriptionInput struct {
	ServiceName *string
	Price       *int
	EndDate     *time.Time
}

type ListParams struct {
	Limit  int32
	Offset int32
}

type CostFilter struct {
	StartPeriod time.Time // "01-2025"
	EndPeriod   time.Time // "12-2025"
	UserID      *uuid.UUID
	ServiceName *string
}

type TotalCostResult struct {
	TotalCost int64
	Count     int64
}

// ParseMonthYear парсит "07-2025" в time.Time
func ParseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected MM-YYYY: %w", err)
	}
	return t, nil
}

// FormatMonthYear форматирует time.Time в "07-2025"
func FormatMonthYear(t time.Time) string {
	return t.Format("01-2006")
}
