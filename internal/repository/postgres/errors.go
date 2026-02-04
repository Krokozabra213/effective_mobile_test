package postgres

import "errors"

var (
	ErrNotFound = errors.New("subscription not found")
	ErrInternal = errors.New("internal repository error")
)
