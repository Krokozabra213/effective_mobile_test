package handler

import (
	"fmt"
	"time"
)

// ParseMonthYear парсит "07-2025" в time.Time
func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected MM-YYYY: %w", err)
	}
	return t, nil
}

// FormatMonthYear форматирует time.Time в "07-2025"
func formatMonthYear(t time.Time) string {
	return t.Format("01-2006")
}
