package model

import (
	"github.com/google/uuid"
	"time"
)

// запись в бд
type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// структура для данных, получаемых в HTTP-запросе POST
type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

// запрос на обновление (PUT/PATCH)
type UpdateSubscriptionRequest struct {
    ServiceName *string `json:"service_name,omitempty"`
    Price       *int    `json:"price,omitempty"`
    StartDate   *string `json:"start_date,omitempty"`
    EndDate     *string `json:"end_date,omitempty"`
}

// сбор параметров аналитики из URL
type CostAnalyticsRequest struct {
	UserID       string `json:"user_id"`
	ServiceName  string `json:"service_name"`
	StartDateStr string `json:"start_date_from"`
	EndDateStr   string `json:"start_date_to"`
}
