// Package handler provides HTTP handlers for API.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Krokozabra213/effective_mobile/docs"
	"github.com/Krokozabra213/effective_mobile/internal/domain"
	"github.com/google/uuid"

	_ "github.com/Krokozabra213/effective_mobile/docs"
)

// Business defines business layer interface.
type Business interface {
	CreateSubscription(ctx context.Context, input *domain.CreateSubscriptionInput) (*domain.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int64) (*domain.Subscription, error)
	ListSubscriptions(ctx context.Context, params domain.ListParams) ([]domain.Subscription, error)
	ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID, params domain.ListParams) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.Subscription, error)
	DeleteSubscription(ctx context.Context, id int64) error
	CalculateTotalCost(ctx context.Context, filter domain.CostFilter) (domain.TotalCost, error)
}

// Handler handles HTTP requests.
type Handler struct {
	business Business
}

// New creates a new Handler and registers routes.
func New(mux *http.ServeMux, business Business) *Handler {
	h := &Handler{
		business: business,
	}

	// Subscriptions CRUD
	mux.HandleFunc("POST /subscriptions", h.CreateSubscription)
	mux.HandleFunc("GET /subscriptions/{id}", h.GetSubscriptionByID)
	mux.HandleFunc("GET /subscriptions", h.ListSubscriptions)
	mux.HandleFunc("PATCH /subscriptions/{id}", h.UpdateSubscription)
	mux.HandleFunc("DELETE /subscriptions/{id}", h.DeleteSubscription)
	mux.HandleFunc("GET /subscriptions/cost", h.CalculateTotalCost)

	// User subscriptions
	mux.HandleFunc("GET /users/{user_id}/subscriptions", h.ListSubscriptionsByUserID)

	// Swagger
	mux.HandleFunc("GET /swagger/doc.json", h.SwaggerJSON)
	mux.HandleFunc("GET /swagger/", h.SwaggerUI)

	return h
}

// SwaggerUI отдаёт HTML страницу Swagger UI
func (h *Handler) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Subscription API</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/swagger/doc.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout"
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// SwaggerJSON отдаёт swagger.json
func (h *Handler) SwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
}

// CreateSubscription создаёт новую подписку
// @Summary      Создать подписку
// @Description  Создаёт новую подписку для пользователя
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request  body      CreateSubscriptionRequest  true  "Данные подписки"
// @Success      201      {object}  SubscriptionResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidBody)
		return
	}
	if err := req.Validate(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := parseMonthYear(*req.EndDate)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
			return
		}
		endDate = &parsed
	}

	input := domain.NewCreateSubscriptionInput(req.ServiceName, req.Price, req.UserID, startDate, endDate)

	sub, err := h.business.CreateSubscription(r.Context(), &input)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}
	resp := h.toSubscriptionResponse(sub)

	h.respondJSON(w, http.StatusCreated, &resp)
}

// GetSubscriptionByID получает подписку по ID
// @Summary      Получить подписку
// @Description  Возвращает подписку по её идентификатору
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      int  true  "ID подписки"
// @Success      200  {object}  SubscriptionResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidIDFormat)
		return
	}

	if id <= 0 {
		h.respondError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	sub, err := h.business.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, h.toSubscriptionResponse(sub))
}

// ListSubscriptions возвращает список всех подписок
// @Summary      Список подписок
// @Description  Возвращает список всех подписок с пагинацией
// @Tags         subscriptions
// @Produce      json
// @Param        limit   query     int  false  "Лимит (по умолчанию 10, макс 100)"
// @Param        offset  query     int  false  "Смещение (по умолчанию 0)"
// @Success      200     {object}  ListSubscriptionsResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /subscriptions [get]
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	params := h.parsePagination(r)

	subs, err := h.business.ListSubscriptions(r.Context(), params)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	response := map[string]any{
		"subscriptions": h.toSubscriptionListResponse(subs),
	}

	h.respondJSON(w, http.StatusOK, response)
}

// ListSubscriptionsByUserID возвращает подписки пользователя
// @Summary      Подписки пользователя
// @Description  Возвращает список подписок конкретного пользователя
// @Tags         users
// @Produce      json
// @Param        user_id  path      string  true  "UUID пользователя"
// @Param        limit    query     int     false  "Лимит (по умолчанию 10, макс 100)"
// @Param        offset   query     int     false  "Смещение (по умолчанию 0)"
// @Success      200      {object}  ListSubscriptionsResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /users/{user_id}/subscriptions [get]
func (h *Handler) ListSubscriptionsByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidUserIDFormat)
		return
	}

	params := h.parsePagination(r)

	subs, err := h.business.ListSubscriptionsByUserID(r.Context(), userID, params)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	response := map[string]any{
		"subscriptions": h.toSubscriptionListResponse(subs),
	}

	h.respondJSON(w, http.StatusOK, response)
}

// UpdateSubscription обновляет подписку
// @Summary      Обновить подписку
// @Description  Частично обновляет подписку по ID. Все поля опциональны.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id       path      int                        true  "ID подписки"
// @Param        request  body      UpdateSubscriptionRequest   true  "Поля для обновления"
// @Success      200      {object}  SubscriptionResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /subscriptions/{id} [patch]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidIDFormat)
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidBody)
		return
	}

	input := domain.UpdateSubscriptionInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
	}

	if req.EndDate != nil {
		parsed, err := parseMonthYear(*req.EndDate)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
			return
		}
		input.EndDate = &parsed
	}

	sub, err := h.business.UpdateSubscription(r.Context(), id, input)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, h.toSubscriptionResponse(sub))
}

// DeleteSubscription удаляет подписку
// @Summary      Удалить подписку
// @Description  Удаляет подписку по ID
// @Tags         subscriptions
// @Param        id   path  int  true  "ID подписки"
// @Success      204  "Подписка удалена"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r, "id")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidIDFormat)
		return
	}

	if err := h.business.DeleteSubscription(r.Context(), id); err != nil {
		h.handleBusinessError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CalculateTotalCost рассчитывает суммарную стоимость подписок
// @Summary      Рассчитать стоимость
// @Description  Рассчитывает суммарную стоимость подписок за период с опциональной фильтрацией по пользователю и сервису
// @Tags         subscriptions
// @Produce      json
// @Param        start_period   query     string  true   "Начало периода (MM-YYYY)"  example(01-2024)
// @Param        end_period     query     string  true   "Конец периода (MM-YYYY)"   example(12-2024)
// @Param        user_id        query     string  false  "UUID пользователя"
// @Param        service_name   query     string  false  "Название сервиса"
// @Success      200            {object}  TotalCostResponse
// @Failure      400            {object}  ErrorResponse
// @Failure      500            {object}  ErrorResponse
// @Router       /subscriptions/cost [get]
func (h *Handler) CalculateTotalCost(w http.ResponseWriter, r *http.Request) {
	startPeriod, err := parseMonthYear(r.URL.Query().Get("start_period"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
		return
	}

	endPeriod, err := parseMonthYear(r.URL.Query().Get("end_period"))
	if err != nil {
		h.respondError(w, http.StatusBadRequest, ErrInvalidDate)
		return
	}

	filter := domain.CostFilter{
		StartPeriod: startPeriod,
		EndPeriod:   endPeriod,
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, ErrInvalidUserIDFormat)
			return
		}
		filter.UserID = &userID
	}

	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	result, err := h.business.CalculateTotalCost(r.Context(), filter)
	if err != nil {
		h.handleBusinessError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, NewTotalCostResponse(result.TotalCost, result.Count))
}
