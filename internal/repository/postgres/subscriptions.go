package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Krokozabra213/effective_mobile/internal/domain"
	sqlc "github.com/Krokozabra213/effective_mobile/internal/repository/postgres/queries"
	"github.com/google/uuid"
)

// CreateSubscription создаёт подписку
func (r *PostgresRepository) CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error) {
	const op = "repository.CreateSubscription"
	log := slog.With(slog.String("op", op))

	result, err := r.Queries.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
		ServiceName: input.ServiceName,
		Price:       int32(input.Price),
		UserID:      input.UserID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	})
	if err != nil {
		log.Error("failed to create subscription", slog.String("error", err.Error()))
		return nil, r.handleError(err)
	}

	return r.toDomain(&result), nil
}

// GetSubscriptionByID получает подписку по ID
func (r *PostgresRepository) GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error) {
	const op = "repository.GetSubscriptionByID"
	log := slog.With(slog.String("op", op), slog.Int64("id", id))

	result, err := r.Queries.GetSubscriptionByID(ctx, id)
	if err != nil {
		log.Error("failed to get subscription", slog.String("error", err.Error()))
		return nil, r.handleError(err)
	}

	return r.toDomain(&result), nil
}

// ListSubscriptions возвращает список подписок с пагинацией
func (r *PostgresRepository) ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error) {
	const op = "repository.ListSubscriptions"
	log := slog.With(slog.String("op", op))

	results, err := r.Queries.ListSubscriptions(ctx, sqlc.ListSubscriptionsParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		log.Error("failed to list subscriptions", slog.String("error", err.Error()))
		return nil, r.handleError(err)
	}

	subs := make([]domain.Subscription, len(results))
	for i, result := range results {
		subs[i] = *r.toDomain(&result)
	}

	return subs, nil
}

// ListSubscriptionsByUserID возвращает подписки пользователя
func (r *PostgresRepository) ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error) {
	const op = "repository.ListSubscriptionsByUserID"
	log := slog.With(slog.String("op", op), slog.String("user_id", userID.String()))

	results, err := r.Queries.ListSubscriptionsByUserID(ctx, sqlc.ListSubscriptionsByUserIDParams{
		UserID: userID,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		log.Error("failed to list user subscriptions", slog.String("error", err.Error()))
		return nil, r.handleError(err)
	}

	subs := make([]domain.Subscription, len(results))
	for i, result := range results {
		subs[i] = *r.toDomain(&result)
	}

	return subs, nil
}

// UpdateSubscription обновляет подписку
func (r *PostgresRepository) UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error) {
	const op = "repository.UpdateSubscription"
	log := slog.With(slog.String("op", op), slog.Int64("id", id))

	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.ServiceName != nil {
		setParts = append(setParts, fmt.Sprintf("service_name = $%d", argIndex))
		args = append(args, *input.ServiceName)
		argIndex++
	}

	if input.Price != nil {
		setParts = append(setParts, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *input.Price)
		argIndex++
	}

	if input.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIndex))
		args = append(args, *input.EndDate)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetSubscriptionByID(ctx, id)
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE subscriptions
		SET %s
		WHERE id = $%d
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at
	`, strings.Join(setParts, ", "), argIndex)

	var result sqlc.Subscription
	err := r.DB.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.ServiceName,
		&result.Price,
		&result.UserID,
		&result.StartDate,
		&result.EndDate,
		&result.CreatedAt,
	)
	if err != nil {
		log.Error("failed to update subscription", slog.String("error", err.Error()))
		return nil, r.handleError(err)
	}

	return r.toDomain(&result), nil
}

// DeleteSubscription удаляет подписку
func (r *PostgresRepository) DeleteSubscription(ctx context.Context, id int64) error {
	const op = "repository.DeleteSubscription"
	log := slog.With(slog.String("op", op), slog.Int64("id", id))

	rowsAffected, err := r.Queries.DeleteSubscription(ctx, id)
	if err != nil {
		log.Error("failed to delete subscription", slog.String("error", err.Error()))
		return r.handleError(err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// CalculateTotalCost подсчитывает суммарную стоимость подписок за период
func (r *PostgresRepository) CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (*domain.TotalCostResult, error) {
	const op = "repository.CalculateTotalCost"
	log := slog.With(slog.String("op", op))

	// Конец периода = последний день месяца
	endPeriod := filter.EndPeriod.AddDate(0, 1, -1)

	// Динамический запрос с опциональными фильтрами
	query := `
		SELECT COALESCE(SUM(price), 0)::BIGINT AS total_cost, COUNT(*)::BIGINT AS count
		FROM subscriptions
		WHERE start_date <= $1
		  AND (end_date IS NULL OR end_date >= $2)
	`
	args := []interface{}{endPeriod, filter.StartPeriod}
	argIndex := 3

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, *filter.ServiceName)
	}

	var result domain.TotalCostResult
	if err := r.DB.QueryRow(ctx, query, args...).Scan(&result.TotalCost, &result.Count); err != nil {
		log.Error("failed to calculate total cost", slog.String("error", err.Error()))
		return nil, r.handleError(err)
	}

	return &result, nil
}

// toDomain конвертирует sqlc модель в domain
func (r *PostgresRepository) toDomain(s *sqlc.Subscription) *domain.Subscription {
	return &domain.Subscription{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       int(s.Price),
		UserID:      s.UserID,
		StartDate:   s.StartDate,
		EndDate:     s.EndDate,
		CreatedAt:   s.CreatedAt,
	}
}
