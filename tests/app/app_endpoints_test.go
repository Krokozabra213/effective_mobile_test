package app_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "github.com/Krokozabra213/effective_mobile/internal/delivery/http"
	"github.com/Krokozabra213/effective_mobile/internal/domain"
	"github.com/Krokozabra213/effective_mobile/tests/app/suite"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func TestCreateSubscription(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	userID := uuid.New()
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Netflix",
		"price":        999,
		"user_id":      userID.String(),
		"start_date":   "01-2024",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var sub handler.SubscriptionResponse
	err = resp.JSON(&sub)
	require.NoError(t, err)

	assert.Equal(t, "Netflix", sub.ServiceName)
	assert.Equal(t, int32(999), sub.Price)
	assert.Equal(t, userID, sub.UserID)
	assert.NotZero(t, sub.ID)
	assert.NotZero(t, sub.CreatedAt)
}

func TestCreateSubscription_WithEndDate(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Spotify",
		"price":        169,
		"user_id":      uuid.New().String(),
		"start_date":   "01-2024",
		"end_date":     "12-2024",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var sub handler.SubscriptionResponse
	err = resp.JSON(&sub)
	require.NoError(t, err)

	assert.NotNil(t, sub.EndDate)
}

func TestCreateSubscription_InvalidPrice(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Test",
		"price":        -100,
		"user_id":      uuid.New().String(),
		"start_date":   "01-2024",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp ErrorResponse
	err = resp.JSON(&errResp)
	require.NoError(t, err)

	assert.NotEmpty(t, errResp.Error)
}

func TestCreateSubscription_InvalidDateFormat(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Test",
		"price":        100,
		"user_id":      uuid.New().String(),
		"start_date":   "2024-01-01", // неправильный формат
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetSubscriptionByID(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	// Создаём подписку
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "YouTube Premium",
		"price":        299,
		"user_id":      uuid.New().String(),
		"start_date":   "06-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created domain.Subscription
	err = resp.JSON(&created)
	require.NoError(t, err)

	// Получаем по ID
	path := fmt.Sprintf("/subscriptions/%d", created.ID)
	resp, err = st.HTTPClient.GET(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var fetched domain.Subscription
	err = resp.JSON(&fetched)
	require.NoError(t, err)

	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.ServiceName, fetched.ServiceName)
	assert.Equal(t, created.Price, fetched.Price)
}

func TestGetSubscriptionByID_NotFound(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions/999999")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetSubscriptionByID_InvalidID(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions/abc")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestListSubscriptions(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	userID := uuid.New()

	// Создаём 3 подписки
	for i := 1; i <= 3; i++ {
		resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
			"service_name": fmt.Sprintf("Service_%d", i),
			"price":        100 * i,
			"user_id":      userID.String(),
			"start_date":   "01-2024",
		})
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Subscriptions []handler.SubscriptionResponse `json:"subscriptions"`
	}
	err = resp.JSON(&result)
	require.NoError(t, err)

	fmt.Println()

	assert.Len(t, result.Subscriptions, 3)
}

func TestListSubscriptions_Empty(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Subscriptions []handler.SubscriptionResponse `json:"subscriptions"`
	}
	err = resp.JSON(&result)
	require.NoError(t, err)

	assert.Empty(t, result.Subscriptions)
}

func TestListSubscriptionsByUserID(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	userID := uuid.New()
	otherUserID := uuid.New()

	// Подписки пользователя
	for i := 1; i <= 2; i++ {
		resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
			"service_name": "UserService",
			"price":        200 * i,
			"user_id":      userID.String(),
			"start_date":   "03-2024",
		})
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// Подписка другого пользователя
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "OtherService",
		"price":        500,
		"user_id":      otherUserID.String(),
		"start_date":   "03-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Получаем подписки только нужного пользователя
	path := fmt.Sprintf("/users/%s/subscriptions", userID)
	resp, err = st.HTTPClient.GET(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Subscriptions []handler.SubscriptionResponse `json:"subscriptions"`
	}
	err = resp.JSON(&result)
	require.NoError(t, err)

	assert.Len(t, result.Subscriptions, 2)
	for _, sub := range result.Subscriptions {
		assert.Equal(t, userID, sub.UserID)
	}
}

func TestListSubscriptionsByUserID_InvalidUUID(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/users/invalid-uuid/subscriptions")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestUpdateSubscription(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	// Создаём подписку
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Original",
		"price":        100,
		"user_id":      uuid.New().String(),
		"start_date":   "01-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created domain.Subscription
	err = resp.JSON(&created)
	require.NoError(t, err)

	// Обновляем
	path := fmt.Sprintf("/subscriptions/%d", created.ID)
	resp, err = st.HTTPClient.PATCH(ctx, path, map[string]any{
		"service_name": "Updated",
		"price":        200,
		"end_date":     "12-2024",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated handler.SubscriptionResponse
	err = resp.JSON(&updated)
	require.NoError(t, err)

	assert.Equal(t, "Updated", updated.ServiceName)
	assert.Equal(t, int32(200), updated.Price)
	assert.NotNil(t, updated.EndDate)
}

func TestUpdateSubscription_PartialUpdate(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	// Создаём
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "PartialTest",
		"price":        300,
		"user_id":      uuid.New().String(),
		"start_date":   "02-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created domain.Subscription
	err = resp.JSON(&created)
	require.NoError(t, err)

	// Обновляем только цену
	path := fmt.Sprintf("/subscriptions/%d", created.ID)
	resp, err = st.HTTPClient.PATCH(ctx, path, map[string]any{
		"price": 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated handler.SubscriptionResponse
	err = resp.JSON(&updated)
	require.NoError(t, err)

	assert.Equal(t, "PartialTest", updated.ServiceName) // не изменилось
	assert.Equal(t, int32(400), updated.Price)          // изменилось
}

func TestUpdateSubscription_NotFound(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.PATCH(ctx, "/subscriptions/999999", map[string]any{
		"service_name": "Test",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteSubscription(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	// Создаём
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "ToDelete",
		"price":        100,
		"user_id":      uuid.New().String(),
		"start_date":   "01-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created domain.Subscription
	err = resp.JSON(&created)
	require.NoError(t, err)

	// Удаляем
	path := fmt.Sprintf("/subscriptions/%d", created.ID)
	resp, err = st.HTTPClient.DELETE(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Проверяем что удалено
	resp, err = st.HTTPClient.GET(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.DELETE(ctx, "/subscriptions/999999")
	if err != nil {
		t.Fatal(err)
	}

	// Может быть 404 или 204
	assert.Contains(t, []int{http.StatusNotFound, http.StatusNoContent}, resp.StatusCode)
}

func TestCalculateTotalCost(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	userID := uuid.New()

	// Создаём подписки
	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Netflix",
		"price":        1000,
		"user_id":      userID.String(),
		"start_date":   "01-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "Spotify",
		"price":        200,
		"user_id":      userID.String(),
		"start_date":   "03-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = st.HTTPClient.GET(ctx, "/subscriptions/cost?start_period=01-2024&end_period=12-2024")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		TotalCost int64 `json:"total_cost"`
		Count     int64 `json:"count"`
	}
	err = resp.JSON(&result)
	require.NoError(t, err)

	assert.Equal(t, int64(2), result.Count)
	assert.Greater(t, result.TotalCost, int64(0))
}

func TestCalculateTotalCost_FilterByUserID(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	userID := uuid.New()

	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": "UserSpecific",
		"price":        500,
		"user_id":      userID.String(),
		"start_date":   "06-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	path := fmt.Sprintf("/subscriptions/cost?start_period=01-2024&end_period=12-2024&user_id=%s", userID)
	resp, err = st.HTTPClient.GET(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		TotalCost int64 `json:"total_cost"`
		Count     int64 `json:"count"`
	}
	err = resp.JSON(&result)
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Count)
	assert.Equal(t, int64(500), result.TotalCost)
}

func TestCalculateTotalCost_FilterByServiceName(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	serviceName := "UniqueService"

	resp, err := st.HTTPClient.POST(ctx, "/subscriptions", map[string]any{
		"service_name": serviceName,
		"price":        777,
		"user_id":      uuid.New().String(),
		"start_date":   "05-2024",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	path := fmt.Sprintf("/subscriptions/cost?start_period=01-2024&end_period=12-2024&service_name=%s", serviceName)
	resp, err = st.HTTPClient.GET(ctx, path)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		TotalCost int64 `json:"total_cost"`
		Count     int64 `json:"count"`
	}
	err = resp.JSON(&result)
	require.NoError(t, err)

	assert.Equal(t, int64(1), result.Count)
}

func TestCalculateTotalCost_MissingPeriod(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions/cost?end_period=12-2024")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCalculateTotalCost_InvalidDateFormat(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions/cost?start_period=2024-01-01&end_period=2024-12-31")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCalculateTotalCost_InvalidUserID(t *testing.T) {
	ctx, st := suite.New(t)
	t.Cleanup(func() {
		st.CleanupTestData()
	})

	resp, err := st.HTTPClient.GET(ctx, "/subscriptions/cost?start_period=01-2024&end_period=12-2024&user_id=invalid")
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
