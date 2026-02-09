package business

import (
	"context"
	"log/slog"

	"github.com/Krokozabra213/effective_mobile/internal/domain"
	"github.com/google/uuid"
)

type BusinessInterface interface {
	CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error)
	ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error)
	ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error)
	DeleteSubscription(ctx context.Context, id int64) error
	CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (domain.TotalCost, error)
}

type SubscriptionProvider interface {
	CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error)
	ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error)
	ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error)
	DeleteSubscription(ctx context.Context, id int64) error
	CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (domain.TotalCost, error)
}

// Business contains the core business logic and dependencies.
type Business struct {
	log  *slog.Logger
	repo SubscriptionProvider
}

// New creates a new Business instance with the provided dependencies.
func New(log *slog.Logger, repo SubscriptionProvider) *Business {
	return &Business{
		log:  log,
		repo: repo,
	}
}

var _ BusinessInterface = (*Business)(nil)
