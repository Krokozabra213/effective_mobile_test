package postgres

import (
	"context"
	"errors"

	"github.com/Krokozabra213/effective_mobile/internal/domain"
	sqlc "github.com/Krokozabra213/effective_mobile/internal/repository/postgres/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)



type SubscriptionProvider interface {
    CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error)
    GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error)
    ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error)
    ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error)
    UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error)
    DeleteSubscription(ctx context.Context, id int64) error
    CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (*domain.TotalCostResult, error)
}

var _ SubscriptionProvider = (*PostgresRepository)(nil)

type PostgresRepository struct {
	DB      sqlc.DBTX
	Queries sqlc.Querier
}

func NewRepository(db sqlc.DBTX) *PostgresRepository {
	return &PostgresRepository{
		DB:      db,
		Queries: sqlc.New(db),
	}
}

func (r *PostgresRepository) handleError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return ErrInternal
}
