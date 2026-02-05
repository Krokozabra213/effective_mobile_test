package business

import (
	"errors"

	repository "github.com/Krokozabra213/effective_mobile/internal/repository/postgres"
)

var (
	ErrNotFound = errors.New("subscription not found")
	ErrInternal = errors.New("internal error")
)

func (b *Business) mapError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return ErrInternal
}
