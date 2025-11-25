package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"effective-mobile-subscriptions/internal/model"
	"effective-mobile-subscriptions/internal/service"
	"github.com/gorilla/mux"
)

// содержит логику для HTTP-обработки
type SubscriptionHandler struct{ Service *service.SubscriptionService }

func NewSubscriptionHandler(s *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{Service: s}
}

// структуры для корректного отображения ошибок в Swagger
type BadRequestResponse struct {
	Error string `json:"error" example:"incorrect format user_id (expected UUID)"`
}
type SubscriptionNotFoundResponse struct {
	Error string `json:"error" example:"Subscription not found"`
}
type InternalServerErrorResponse struct {
	Error string `json:"error" example:"Internal Server Error"`
}

// отправляет JSON-ответ
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("CRITICAL ERROR: Failed to marshal payload to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// @Summary Создать новую подписку
// @Description Создает новую запись об онлайн-подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.CreateSubscriptionRequest true "Данные новой подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} BadRequestResponse "Некорректный запрос или ошибка валидации (UUID, дата, формат JSON)"
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Invalid request payload: %v", err)
		RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload or malformed JSON"})
		return
	}
	sub, err := h.Service.Create(r.Context(), req)
	if err != nil {
		log.Printf("ERROR: Service failed to create subscription: %v", err)
		RespondServiceError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, sub)
}

// @Summary Получить подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "UUID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} BadRequestResponse "Некорректный формат ID"
// @Failure 404 {object} SubscriptionNotFoundResponse "Подписка не найдена"
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	sub, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		RespondServiceError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, sub)
}

// @Summary Обновить существующую подписку
// @Description Обновляет существующую запись об онлайн-подписке, используя переданные поля.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "UUID подписки"
// @Param subscription body model.UpdateSubscriptionRequest true "Обновленные данные подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} BadRequestResponse "Некорректный запрос, формат ID или ошибка валидации"
// @Failure 404 {object} SubscriptionNotFoundResponse "Подписка не найдена"
// @Failure 500 {object} InternalServerErrorResponse "Ошибка сервиса или БД"
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode request body for update: %v", err)
		RespondJSON(w, http.StatusBadRequest, BadRequestResponse{Error: "Incorrect format JSON"})
		return
	}
	updatedSub, err := h.Service.Update(r.Context(), id, req)
	if err != nil {
		log.Printf("ERROR: Failed to update subscription %s in service: %v", id, err)
		RespondServiceError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, updatedSub)
}

// @Summary Удалить подписку по ID
// @Tags subscriptions
// @Param id path string true "UUID подписки"
// @Success 204 "Подписка успешно удалена (No Content)"
// @Failure 400 {object} BadRequestResponse "Некорректный формат ID"
// @Failure 404 {object} SubscriptionNotFoundResponse "Подписка не найдена"
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	deleted, err := h.Service.Delete(r.Context(), id)
	if err != nil {
		RespondServiceError(w, err)
		return
	}
	if !deleted {
		RespondJSON(w, http.StatusNotFound, map[string]string{"error": "Subscription not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получить список всех подписок
// @Tags subscriptions
// @Produce json
// @Success 200 {array} model.Subscription
// @Failure 500 {object} InternalServerErrorResponse "Ошибка БД/сервиса"
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.Service.List(r.Context())
	if err != nil {
		log.Printf("FATAL ERROR: Service failed to fetch list of subscriptions: %v", err)
		RespondJSON(w, http.StatusInternalServerError, InternalServerErrorResponse{Error: "Internal Server Error"})
		return
	}
	RespondJSON(w, http.StatusOK, subscriptions)
}

type CostAnalyticsResponse struct {
	TotalCost int `json:"total_cost"`
}

// @Summary Подсчет суммарной стоимости подписок по фильтрам
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Фильтр по UUID пользователя"
// @Param service_name query string false "Фильтр по названию подписки"
// @Param start_date_from query string false "Период от (MM-YYYY)"
// @Param start_date_to query string false "Период до (MM-YYYY)"
// @Success 200 {object} CostAnalyticsResponse
// @Failure 400 {object} BadRequestResponse "Ошибка валидации параметров запроса (UUID, дата)"
// @Router /subscriptions/analytics [get]
func (h *SubscriptionHandler) GetCostAnalytics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	req := model.CostAnalyticsRequest{
		UserID:       query.Get("user_id"),
		ServiceName:  query.Get("service_name"),
		StartDateStr: query.Get("start_date_from"),
		EndDateStr:   query.Get("start_date_to"),
	}
	totalCost, err := h.Service.GetCostAnalytics(r.Context(), req)
	if err != nil {
		log.Printf("WARN: Analytics request validation error: %v", err)
		RespondServiceError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, CostAnalyticsResponse{TotalCost: totalCost})
}

func RespondServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		RespondJSON(w, http.StatusBadRequest, BadRequestResponse{Error: UserFacingErrorMessage(err)})
	case errors.Is(err, service.ErrNotFound):
		RespondJSON(w, http.StatusNotFound, SubscriptionNotFoundResponse{Error: "Subscription not found"})
	default:
		RespondJSON(w, http.StatusInternalServerError, InternalServerErrorResponse{Error: "Internal Server Error"})
	}
}

func UserFacingErrorMessage(err error) string {
	const delimiter = ": "
	msg := err.Error()
	prefix := service.ErrValidation.Error() + delimiter
	if strings.HasPrefix(msg, prefix) {
		return msg[len(prefix):]
	}
	return msg
}
