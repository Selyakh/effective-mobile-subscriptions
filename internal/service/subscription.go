package service

import (
	"context"
	"fmt"
	"log"
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
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("incorrect format user_id (expected UUID): %w", err)
	}
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("incorrect format start_date (expected MM-YYYY): %w", err)
	}
	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsedEndDate, err := time.Parse("01-2006", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("incorrect format end_date (expected MM-YYYY): %w", err)
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
		return nil, fmt.Errorf("incorrect format ID: %w", err)
	}
	sub, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("ERROR: GetByID failed to fetch subscription for ID %s from repository: %v", idStr, err)
		return nil, fmt.Errorf("service error when receiving a subscription: %w", err)
	}
	return sub, nil
}

// обновить существующую подписку (только переданные поля)
func (s *SubscriptionService) Update(ctx context.Context, id string, req model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	subID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("incorrect format subscription ID (expected UUID): %w", err)
	}
	existingSub, err := s.Repo.GetByID(ctx, subID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch existing subscription %s from repository: %v", id, err)
		return nil, fmt.Errorf("failed to retrieve subscription for update: %w", err)
	}
	if existingSub == nil {
		return nil, fmt.Errorf("subscription not found")
	}
	if req.ServiceName != nil {
		existingSub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		existingSub.Price = *req.Price
	}
	if req.StartDate != nil {
		startDate, err := time.Parse("01-2006", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("incorrect format start_date (expected MM-YYYY): %w", err)
		}
		existingSub.StartDate = startDate
	}
	if req.EndDate != nil {
		if *req.EndDate != "" {
			parsedEndDate, err := time.Parse("01-2006", *req.EndDate)
			if err != nil {
				return nil, fmt.Errorf("incorrect format end_date (expected MM-YYYY): %w", err)
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
		return false, fmt.Errorf("incorrect format ID: %w", err)
	}
	deleted, err := s.Repo.Delete(ctx, id)
	if err != nil {
		log.Printf("ERROR: Delete failed to remove subscription for ID %s from repository: %v", idStr, err)
		return false, fmt.Errorf("service error when deleting a subscription: %w", err)
	}
	return deleted, nil
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
	if req.StartDateStr != "" {
		_, err := time.Parse("01-2006", req.StartDateStr)
		if err != nil {
			return 0, fmt.Errorf("incorrect format start_date_from (expected MM-YYYY)")
		}
		req.StartDateStr = req.StartDateStr[3:] + "-" + req.StartDateStr[:2] + "-01"
	}
	if req.EndDateStr != "" {
		_, err := time.Parse("01-2006", req.EndDateStr)
		if err != nil {
			return 0, fmt.Errorf("incorrect format start_date_to (expected MM-YYYY)")
		}
		req.EndDateStr = req.EndDateStr[3:] + "-" + req.EndDateStr[:2] + "-01"
	}
	if req.UserID != "" {
		if _, err := uuid.Parse(req.UserID); err != nil {
			return 0, fmt.Errorf("incorrect format user_id (expected UUID)")
		}
	}
	totalCost, err := s.Repo.GetTotalCost(ctx, req)
	if err != nil {
		log.Printf("ERROR: GetCostAnalytics failed to execute total cost query in repository: %v", err)
		return 0, fmt.Errorf("service error while receiving analytics: %w", err)
	}
	return totalCost, nil
}
