package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"effective-mobile-subscriptions/internal/model"
	"effective-mobile-subscriptions/internal/repository"
	"github.com/google/uuid"
)

// определить методы бизнес-логики
type SubscriptionService struct {
	Repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{Repo: repo}
}

// создать подписку
func (s *SubscriptionService) Create(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error) {
	if err := ValidateCreateRequest(req); err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ValidationError("incorrect format user_id (expected UUID)")
	}
	startDate, err := ParseMonthYear("start_date", req.StartDate)
	if err != nil {
		return nil, err
	}
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsedEndDate, err := ParseMonthYear("end_date", *req.EndDate)
		if err != nil {
			return nil, err
		}
		endDate = &parsedEndDate
	}
	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	if err := s.Repo.Create(ctx, sub); err != nil {
		log.Printf("ERROR: Failed to create subscription in repository: %v", err)
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}
	return sub, nil
}

// получить подписку по её ID
func (s *SubscriptionService) GetByID(ctx context.Context, idStr string) (*model.Subscription, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, ValidationError("incorrect format ID (expected UUID)")
	}
	sub, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("ERROR: GetByID failed to fetch subscription for ID %s from repository: %v", idStr, err)
		return nil, fmt.Errorf("service error when receiving a subscription: %w", err)
	}
	if sub == nil {
		return nil, ErrNotFound
	}
	return sub, nil
}

// обновить существующую подписку (только переданные поля)
func (s *SubscriptionService) Update(ctx context.Context, id string, req model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	if err := ValidateUpdateRequest(req); err != nil {
		return nil, err
	}
	subID, err := uuid.Parse(id)
	if err != nil {
		return nil, ValidationError("incorrect format subscription ID (expected UUID)")
	}
	existingSub, err := s.Repo.GetByID(ctx, subID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch existing subscription %s from repository: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve subscription for update: %w", err)
	}
	if existingSub == nil {
		return nil, ErrNotFound
	}
	if req.ServiceName != nil {
		existingSub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		existingSub.Price = *req.Price
	}
	if req.StartDate != nil {
		startDate, err := ParseMonthYear("start_date", *req.StartDate)
		if err != nil {
			return nil, err
		}
		existingSub.StartDate = startDate
	}
	if req.EndDate != nil {
		if *req.EndDate != "" {
			parsedEndDate, err := ParseMonthYear("end_date", *req.EndDate)
			if err != nil {
				return nil, err
			}
			existingSub.EndDate = &parsedEndDate
		} else {
			existingSub.EndDate = nil
		}
	}
	if err := s.Repo.Update(ctx, existingSub); err != nil {
		log.Printf("ERROR: Failed to update subscription %s in repository: %v", id, err)
		return nil, fmt.Errorf("failed to save updated subscription: %w", err)
	}
	return existingSub, nil
}

// удалить подписку по ID
func (s *SubscriptionService) Delete(ctx context.Context, idStr string) (bool, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return false, ValidationError("incorrect format ID (expected UUID)")
	}
	deleted, err := s.Repo.Delete(ctx, id)
	if err != nil {
		log.Printf("ERROR: Delete failed to remove subscription for ID %s from repository: %v", idStr, err)
		return false, fmt.Errorf("service error when deleting a subscription: %w", err)
	}
	if !deleted {
		return false, ErrNotFound
	}
	return true, nil
}

// получить все подписки
func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	subscriptions, err := s.Repo.List(ctx)
	if err != nil {
		log.Printf("ERROR: List failed to retrieve subscriptions from repository: %v", err)
		return nil, fmt.Errorf("service error while retrieving list: %w", err)
	}
	return subscriptions, nil
}

// получить суммарную стоимость по фильтрам
func (s *SubscriptionService) GetCostAnalytics(ctx context.Context, req model.CostAnalyticsRequest) (int, error) {
	filters := model.CostAnalyticsRequest{
		UserID:      req.UserID,
		ServiceName: req.ServiceName,
	}
	var startDateParsed *time.Time
	var endDateParsed *time.Time
	if req.StartDateStr != "" {
		startDate, err := ParseMonthYear("start_date_from", req.StartDateStr)
		if err != nil {
			return 0, err
		}
		startDateParsed = &startDate
		filters.StartDateStr = startDate.Format("2006-01-02")
	}
	if req.EndDateStr != "" {
		endDate, err := ParseMonthYear("start_date_to", req.EndDateStr)
		if err != nil {
			return 0, err
		}
		endDateParsed = &endDate
		filters.EndDateStr = endDate.Format("2006-01-02")
	}
	if startDateParsed != nil && endDateParsed != nil && startDateParsed.After(*endDateParsed) {
		return 0, ValidationError("start_date_from cannot be after start_date_to")
	}
	if req.UserID != "" {
		if _, err := uuid.Parse(req.UserID); err != nil {
			return 0, ValidationError("incorrect format user_id (expected UUID)")
		}
	}
	totalCost, err := s.Repo.GetTotalCost(ctx, filters)
	if err != nil {
		log.Printf("ERROR: GetCostAnalytics failed to execute total cost query in repository: %v", err)
		return 0, fmt.Errorf("service error while receiving analytics: %w", err)
	}
	return totalCost, nil
}

const monthYearLayout = "01-2006"

func ParseMonthYear(fieldName, value string) (time.Time, error) {
	parsed, err := time.Parse(monthYearLayout, value)
	if err != nil {
		return time.Time{}, ValidationError(fmt.Sprintf("incorrect format %s (expected MM-YYYY)", fieldName))
	}
	return parsed, nil
}

func ValidateCreateRequest(req model.CreateSubscriptionRequest) error {
	if strings.TrimSpace(req.ServiceName) == "" {
		return ValidationError("service_name is required")
	}
	if req.Price <= 0 {
		return ValidationError("price must be greater than zero")
	}
	if strings.TrimSpace(req.UserID) == "" {
		return ValidationError("user_id is required")
	}
	if strings.TrimSpace(req.StartDate) == "" {
		return ValidationError("start_date is required")
	}
	return nil
}

func ValidateUpdateRequest(req model.UpdateSubscriptionRequest) error {
	if req.ServiceName == nil && req.Price == nil && req.StartDate == nil && req.EndDate == nil {
		return ValidationError("at least one field must be provided for update")
	}
	if req.ServiceName != nil && strings.TrimSpace(*req.ServiceName) == "" {
		return ValidationError("service_name cannot be empty")
	}
	if req.Price != nil && *req.Price <= 0 {
		return ValidationError("price must be greater than zero")
	}
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) == "" {
		return ValidationError("start_date cannot be empty")
	}
	return nil
}
