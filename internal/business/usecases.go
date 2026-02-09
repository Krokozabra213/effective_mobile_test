package business

import (
	"context"
	"log/slog"
	"time"

	"github.com/Krokozabra213/effective_mobile/internal/domain"
	"github.com/google/uuid"
)

// CreateSubscription создаёт новую подписку
func (b *Business) CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error) {
	const op = "business.CreateSubscription"
	log := b.log.With(slog.String("op", op), slog.String("user_id", input.UserID.String()))
	log.Info("process started")

	sub, err := b.repo.CreateSubscription(ctx, input)
	if err != nil {
		log.Error("failed to create subscription", slog.String("error", err.Error()))
		return nil, b.mapError(err)
	}

	log.Info("subscription created", slog.Int64("id", sub.ID))
	return sub, nil
}

// GetSubscriptionByID получает подписку по ID
func (b *Business) GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error) {
	const op = "business.GetSubscriptionByID"
	log := b.log.With(slog.String("op", op), slog.Int64("subscription_id", id))
	log.Info("process started")

	sub, err := b.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		log.Error("failed to get subscription", slog.String("error", err.Error()))
		return nil, b.mapError(err)
	}

	log.Info("success")
	return sub, nil
}

// ListSubscriptions возвращает список подписок
func (b *Business) ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error) {
	const op = "business.ListSubscriptions"
	start := time.Now()

	log := b.log.With(
		slog.String("op", op),
		slog.Int("limit", int(params.Limit)),
		slog.Int("offset", int(params.Offset)),
	)
	log.Info("process started")

	subs, err := b.repo.ListSubscriptions(ctx, params)
	if err != nil {
		log.Error("failed to list subscriptions", slog.String("error", err.Error()))
		return nil, b.mapError(err)
	}

	log.Info("success", slog.Int("count", len(subs)), slog.Duration("duration", time.Since(start)))
	return subs, nil
}

// ListSubscriptionsByUserID возвращает список подписок для user
func (b *Business) ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error) {
	const op = "business.ListSubscriptionsByUserID"
	start := time.Now()

	log := b.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.Int("limit", int(params.Limit)),
		slog.Int("offset", int(params.Offset)),
	)
	log.Info("process started")

	subs, err := b.repo.ListSubscriptionsByUserID(ctx, userID, params)
	if err != nil {
		log.Error("failed to list subscriptions", slog.String("error", err.Error()))
		return nil, b.mapError(err)
	}

	log.Info("success", slog.Int("count", len(subs)), slog.Duration("duration", time.Since(start)))
	return subs, nil
}

// UpdateSubscription обновляет подписку
func (b *Business) UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error) {
	const op = "business.UpdateSubscription"
	log := b.log.With(slog.String("op", op), slog.Int64("subscription_id", id))
	log.Info("process started")

	sub, err := b.repo.UpdateSubscription(ctx, id, input)
	if err != nil {
		log.Error("failed to update subscription", slog.String("error", err.Error()))
		return nil, b.mapError(err)
	}

	log.Info("success", slog.Int64("subscription_id", sub.ID))
	return sub, nil
}

// DeleteSubscription удаляет подписку
func (b *Business) DeleteSubscription(ctx context.Context, id int64) error {
	const op = "business.DeleteSubscription"
	log := b.log.With(slog.String("op", op), slog.Int64("subscription_id", id))
	log.Info("process started")

	err := b.repo.DeleteSubscription(ctx, id)
	if err != nil {
		log.Error("failed to delete subscription", slog.String("error", err.Error()))
		return b.mapError(err)
	}

	log.Info("success")
	return nil
}

// CalculateTotalCost подсчитывает суммарную стоимость подписок
func (b *Business) CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (domain.TotalCost, error) {
	const op = "business.CalculateTotalCost"
	start := time.Now()
	log := b.log.With(slog.String("op", op), slog.Time("start", filter.StartPeriod), slog.Time("end", filter.EndPeriod))
	log.Info("process started")

	result, err := b.repo.CalculateTotalCost(ctx, filter)
	if err != nil {
		log.Error("failed to calculate total cost", slog.String("error", err.Error()))
		return domain.TotalCost{}, b.mapError(err)
	}

	log.Debug("success", slog.Int64("total_cost", result.TotalCost), slog.Int64("count", result.Count),
		slog.Duration("duration", time.Since(start)))
	return result, nil
}
