//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Krokozabra213/effective_mobile/internal/domain"
	repository "github.com/Krokozabra213/effective_mobile/internal/repository/postgres"
)

// Хелперы для создания тестовых данных
func createTestInput(serviceName string, price int, userID uuid.UUID) *domain.CreateSubscriptionInput {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return &domain.CreateSubscriptionInput{
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     nil,
	}
}

func createTestInputWithEndDate(serviceName string, price int, userID uuid.UUID, endDate time.Time) *domain.CreateSubscriptionInput {
	input := createTestInput(serviceName, price, userID)
	input.EndDate = &endDate
	return input
}

func ptr[T any](v T) *T {
	return &v
}

// ==================== CreateSubscription ====================

func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()

	t.Run("success without end_date", func(t *testing.T) {
		cleanup(t)

		userID := uuid.New()
		input := createTestInput("Yandex Plus", 400, userID)

		result, err := testRepo.CreateSubscription(ctx, input)

		require.NoError(t, err)
		assert.NotZero(t, result.ID)
		assert.Equal(t, "Yandex Plus", result.ServiceName)
		assert.Equal(t, 400, result.Price)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, input.StartDate, result.StartDate)
		assert.Nil(t, result.EndDate)
		assert.NotZero(t, result.CreatedAt)
	})

	t.Run("success with end_date", func(t *testing.T) {
		cleanup(t)

		userID := uuid.New()
		endDate := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
		input := createTestInputWithEndDate("Netflix", 800, userID, endDate)

		result, err := testRepo.CreateSubscription(ctx, input)

		require.NoError(t, err)
		assert.NotZero(t, result.ID)
		assert.Equal(t, "Netflix", result.ServiceName)
		assert.Equal(t, 800, result.Price)
		require.NotNil(t, result.EndDate)
		assert.Equal(t, endDate, *result.EndDate)
	})

	t.Run("multiple subscriptions same user", func(t *testing.T) {
		cleanup(t)

		userID := uuid.New()

		sub1, err := testRepo.CreateSubscription(ctx, createTestInput("Service1", 100, userID))
		require.NoError(t, err)

		sub2, err := testRepo.CreateSubscription(ctx, createTestInput("Service2", 200, userID))
		require.NoError(t, err)

		assert.NotEqual(t, sub1.ID, sub2.ID)
		assert.Equal(t, sub1.UserID, sub2.UserID)
	})
}

// ==================== GetSubscriptionByID ====================

func TestGetSubscriptionByID(t *testing.T) {
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		cleanup(t)

		userID := uuid.New()
		created, err := testRepo.CreateSubscription(ctx, createTestInput("Spotify", 169, userID))
		require.NoError(t, err)

		result, err := testRepo.GetSubscriptionByID(ctx, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, "Spotify", result.ServiceName)
		assert.Equal(t, 169, result.Price)
		assert.Equal(t, userID, result.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		cleanup(t)

		result, err := testRepo.GetSubscriptionByID(ctx, 99999)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}

// ==================== ListSubscriptions ====================

func TestListSubscriptions(t *testing.T) {
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		cleanup(t)

		result, err := testRepo.ListSubscriptions(ctx, domain.ListParams{Limit: 10, Offset: 0})

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("returns all subscriptions", func(t *testing.T) {
		cleanup(t)

		// Создаём 3 подписки
		for i := 0; i < 3; i++ {
			_, err := testRepo.CreateSubscription(ctx, createTestInput("Service"+string(rune('A'+i)), 100+i*100, uuid.New()))
			require.NoError(t, err)
		}

		result, err := testRepo.ListSubscriptions(ctx, domain.ListParams{Limit: 10, Offset: 0})

		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("pagination limit", func(t *testing.T) {
		cleanup(t)

		// Создаём 5 подписок
		for i := 0; i < 5; i++ {
			_, err := testRepo.CreateSubscription(ctx, createTestInput("Service", 100, uuid.New()))
			require.NoError(t, err)
		}

		result, err := testRepo.ListSubscriptions(ctx, domain.ListParams{Limit: 2, Offset: 0})

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("pagination offset", func(t *testing.T) {
		cleanup(t)

		// Создаём 5 подписок
		for i := 0; i < 5; i++ {
			_, err := testRepo.CreateSubscription(ctx, createTestInput("Service", 100, uuid.New()))
			require.NoError(t, err)
		}

		result, err := testRepo.ListSubscriptions(ctx, domain.ListParams{Limit: 10, Offset: 3})

		require.NoError(t, err)
		assert.Len(t, result, 2) // 5 - 3 = 2
	})

	t.Run("ordered by created_at desc", func(t *testing.T) {
		cleanup(t)

		sub1, _ := testRepo.CreateSubscription(ctx, createTestInput("First", 100, uuid.New()))
		time.Sleep(10 * time.Millisecond) // небольшая задержка для разных timestamp
		sub2, _ := testRepo.CreateSubscription(ctx, createTestInput("Second", 200, uuid.New()))

		result, err := testRepo.ListSubscriptions(ctx, domain.ListParams{Limit: 10, Offset: 0})

		require.NoError(t, err)
		require.Len(t, result, 2)
		// Последний созданный должен быть первым
		assert.Equal(t, sub2.ID, result[0].ID)
		assert.Equal(t, sub1.ID, result[1].ID)
	})
}

// ==================== ListSubscriptionsByUserID ====================

func TestListSubscriptionsByUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns only user subscriptions", func(t *testing.T) {
		cleanup(t)

		user1 := uuid.New()
		user2 := uuid.New()

		// Подписки user1
		testRepo.CreateSubscription(ctx, createTestInput("Service1", 100, user1))
		testRepo.CreateSubscription(ctx, createTestInput("Service2", 200, user1))

		// Подписка user2
		testRepo.CreateSubscription(ctx, createTestInput("Service3", 300, user2))

		result, err := testRepo.ListSubscriptionsByUserID(ctx, user1, domain.ListParams{Limit: 10, Offset: 0})

		require.NoError(t, err)
		assert.Len(t, result, 2)
		for _, sub := range result {
			assert.Equal(t, user1, sub.UserID)
		}
	})

	t.Run("empty for unknown user", func(t *testing.T) {
		cleanup(t)

		testRepo.CreateSubscription(ctx, createTestInput("Service", 100, uuid.New()))

		unknownUser := uuid.New()
		result, err := testRepo.ListSubscriptionsByUserID(ctx, unknownUser, domain.ListParams{Limit: 10, Offset: 0})

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("pagination works", func(t *testing.T) {
		cleanup(t)

		userID := uuid.New()
		for i := 0; i < 5; i++ {
			testRepo.CreateSubscription(ctx, createTestInput("Service", 100, userID))
		}

		result, err := testRepo.ListSubscriptionsByUserID(ctx, userID, domain.ListParams{Limit: 2, Offset: 1})

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

// ==================== UpdateSubscription ====================

func TestUpdateSubscription(t *testing.T) {
	ctx := context.Background()

	t.Run("update service_name", func(t *testing.T) {
		cleanup(t)

		created, _ := testRepo.CreateSubscription(ctx, createTestInput("OldName", 100, uuid.New()))

		result, err := testRepo.UpdateSubscription(ctx, created.ID, domain.UpdateSubscriptionInput{
			ServiceName: ptr("NewName"),
		})

		require.NoError(t, err)
		assert.Equal(t, "NewName", result.ServiceName)
		assert.Equal(t, 100, result.Price) // не изменился
	})

	t.Run("update price", func(t *testing.T) {
		cleanup(t)

		created, _ := testRepo.CreateSubscription(ctx, createTestInput("Service", 100, uuid.New()))

		result, err := testRepo.UpdateSubscription(ctx, created.ID, domain.UpdateSubscriptionInput{
			Price: ptr(500),
		})

		require.NoError(t, err)
		assert.Equal(t, 500, result.Price)
		assert.Equal(t, "Service", result.ServiceName) // не изменился
	})

	t.Run("update end_date", func(t *testing.T) {
		cleanup(t)

		created, _ := testRepo.CreateSubscription(ctx, createTestInput("Service", 100, uuid.New()))
		assert.Nil(t, created.EndDate)

		endDate := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
		result, err := testRepo.UpdateSubscription(ctx, created.ID, domain.UpdateSubscriptionInput{
			EndDate: &endDate,
		})

		require.NoError(t, err)
		require.NotNil(t, result.EndDate)
		assert.Equal(t, endDate, *result.EndDate)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		cleanup(t)

		created, _ := testRepo.CreateSubscription(ctx, createTestInput("OldService", 100, uuid.New()))

		endDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		result, err := testRepo.UpdateSubscription(ctx, created.ID, domain.UpdateSubscriptionInput{
			ServiceName: ptr("NewService"),
			Price:       ptr(999),
			EndDate:     &endDate,
		})

		require.NoError(t, err)
		assert.Equal(t, "NewService", result.ServiceName)
		assert.Equal(t, 999, result.Price)
		require.NotNil(t, result.EndDate)
		assert.Equal(t, endDate, *result.EndDate)
	})

	t.Run("update nothing returns current state", func(t *testing.T) {
		cleanup(t)

		created, _ := testRepo.CreateSubscription(ctx, createTestInput("Service", 100, uuid.New()))

		result, err := testRepo.UpdateSubscription(ctx, created.ID, domain.UpdateSubscriptionInput{})

		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, created.ServiceName, result.ServiceName)
		assert.Equal(t, created.Price, result.Price)
	})

	t.Run("not found", func(t *testing.T) {
		cleanup(t)

		result, err := testRepo.UpdateSubscription(ctx, 99999, domain.UpdateSubscriptionInput{
			ServiceName: ptr("NewName"),
		})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})
}

// ==================== DeleteSubscription ====================

func TestDeleteSubscription(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		cleanup(t)

		created, _ := testRepo.CreateSubscription(ctx, createTestInput("ToDelete", 100, uuid.New()))

		err := testRepo.DeleteSubscription(ctx, created.ID)
		require.NoError(t, err)

		// Проверяем что удалено
		_, err = testRepo.GetSubscriptionByID(ctx, created.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("not found", func(t *testing.T) {
		cleanup(t)

		err := testRepo.DeleteSubscription(ctx, 99999)

		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("deletes only specified subscription", func(t *testing.T) {
		cleanup(t)

		sub1, _ := testRepo.CreateSubscription(ctx, createTestInput("Keep", 100, uuid.New()))
		sub2, _ := testRepo.CreateSubscription(ctx, createTestInput("Delete", 200, uuid.New()))

		err := testRepo.DeleteSubscription(ctx, sub2.ID)
		require.NoError(t, err)

		// sub1 должен остаться
		result, err := testRepo.GetSubscriptionByID(ctx, sub1.ID)
		require.NoError(t, err)
		assert.Equal(t, "Keep", result.ServiceName)
	})
}

// ==================== CalculateTotalCost ====================

func TestCalculateTotalCost(t *testing.T) {
	ctx := context.Background()

	t.Run("empty database", func(t *testing.T) {
		cleanup(t)

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(0), result.TotalCost)
		assert.Equal(t, int64(0), result.Count)
	})

	t.Run("calculates total for period", func(t *testing.T) {
		cleanup(t)

		// Подписка в периоде
		input1 := createTestInput("Service1", 100, uuid.New())
		input1.StartDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		testRepo.CreateSubscription(ctx, input1)

		// Ещё одна подписка в периоде
		input2 := createTestInput("Service2", 200, uuid.New())
		input2.StartDate = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		testRepo.CreateSubscription(ctx, input2)

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(300), result.TotalCost)
		assert.Equal(t, int64(2), result.Count)
	})

	t.Run("excludes subscriptions outside period", func(t *testing.T) {
		cleanup(t)

		// Подписка началась ПОСЛЕ периода
		input1 := createTestInput("Future", 100, uuid.New())
		input1.StartDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		testRepo.CreateSubscription(ctx, input1)

		// Подписка закончилась ДО периода
		endDate := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
		input2 := createTestInputWithEndDate("Past", 200, uuid.New(), endDate)
		input2.StartDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		testRepo.CreateSubscription(ctx, input2)

		// Подписка в периоде
		input3 := createTestInput("InPeriod", 300, uuid.New())
		input3.StartDate = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		testRepo.CreateSubscription(ctx, input3)

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(300), result.TotalCost)
		assert.Equal(t, int64(1), result.Count)
	})

	t.Run("filters by user_id", func(t *testing.T) {
		cleanup(t)

		user1 := uuid.New()
		user2 := uuid.New()

		testRepo.CreateSubscription(ctx, createTestInput("User1Sub", 100, user1))
		testRepo.CreateSubscription(ctx, createTestInput("User2Sub", 200, user2))

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			UserID:      &user1,
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(100), result.TotalCost)
		assert.Equal(t, int64(1), result.Count)
	})

	t.Run("filters by service_name", func(t *testing.T) {
		cleanup(t)

		testRepo.CreateSubscription(ctx, createTestInput("Yandex Plus", 400, uuid.New()))
		testRepo.CreateSubscription(ctx, createTestInput("Netflix", 800, uuid.New()))
		testRepo.CreateSubscription(ctx, createTestInput("Yandex Plus", 400, uuid.New()))

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			ServiceName: ptr("Yandex Plus"),
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(800), result.TotalCost) // 400 + 400
		assert.Equal(t, int64(2), result.Count)
	})

	t.Run("filters by user_id and service_name", func(t *testing.T) {
		cleanup(t)

		user1 := uuid.New()
		user2 := uuid.New()

		testRepo.CreateSubscription(ctx, createTestInput("Yandex Plus", 400, user1))
		testRepo.CreateSubscription(ctx, createTestInput("Netflix", 800, user1))
		testRepo.CreateSubscription(ctx, createTestInput("Yandex Plus", 400, user2))

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			UserID:      &user1,
			ServiceName: ptr("Yandex Plus"),
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(400), result.TotalCost)
		assert.Equal(t, int64(1), result.Count)
	})

	t.Run("includes active subscriptions (end_date is null)", func(t *testing.T) {
		cleanup(t)

		// Активная подписка (без end_date)
		testRepo.CreateSubscription(ctx, createTestInput("Active", 500, uuid.New()))

		filter := domain.CostFilter{
			StartPeriod: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndPeriod:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		}

		result, err := testRepo.CalculateTotalCost(ctx, filter)

		require.NoError(t, err)
		assert.Equal(t, int64(500), result.TotalCost)
		assert.Equal(t, int64(1), result.Count)
	})
}
