package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"effective-mobile-subscriptions/internal/model"
	"github.com/google/uuid"
)

// определяет методы для работы с бд
type SubscriptionRepository struct {
	DB *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{DB: db}
}

// сохранить новую подписку в бд и возвратить сгенерированные ID и CreatedAt
func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	query := `INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	err := r.DB.QueryRowContext(
		ctx,
		query,
		sub.UserID,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt)
	if err != nil {
		log.Printf("FATAL DB ERROR: Failed to execute INSERT query for new subscription: %v", err)
		return fmt.Errorf("error creating subscription in DB: %w", err)
	}
	return nil
}

// извлечь подписку из бд по её UUID
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date, created_at
	          FROM subscriptions
		      WHERE id = $1`
	sub := &model.Subscription{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.ServiceName,
		&sub.Price,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("ERROR: Failed to execute SELECT query for ID %s: %v", id, err)
		return nil, fmt.Errorf("error receiving subscription from DB: %w", err)
	}
	return sub, nil
}

// обновить существующую подписку в бд
func (r *SubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	query := `UPDATE subscriptions SET
		    service_name = $2,
			price = $3,
			start_date = $4,
			end_date = $5
		    WHERE id = $1
		    RETURNING created_at, user_id`
	var tempUserID uuid.UUID
	err := r.DB.QueryRowContext(
		ctx,
		query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.CreatedAt, &tempUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("update record not found: %w", err)
		}
		log.Printf("ERROR: Failed to execute UPDATE query for ID %s: %v", sub.ID, err)
		return fmt.Errorf("error updating subscription in DB: %w", err)
	}
	return nil
}

// удалить подписку из бд по её ID
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("ERROR: Failed to execute DELETE query for ID %s: %v", id, err)
		return false, fmt.Errorf("error deleting subscription from DB: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("ERROR: Failed to check rows affected after DELETE for ID %s: %v", id, err)
		return false, fmt.Errorf("error checking the number of deleted rows: %w", err)
	}
	return rowsAffected > 0, nil
}

// предоставить весь список существующих подписок
func (r *SubscriptionRepository) List(ctx context.Context) ([]model.Subscription, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date, created_at
		FROM subscriptions
		ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("ERROR: Failed to execute LIST query: %v", err)
		return nil, fmt.Errorf("failed to fetch subscription list from DB: %w", err)
	}
	defer rows.Close()
	subscriptions := make([]model.Subscription, 0)
	for rows.Next() {
		sub := model.Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.ServiceName,
			&sub.Price,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("subscription string scanning error: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating rows: %w", err)
	}
	return subscriptions, nil
}

// подсчитать суммарную стоимость подписок по заданным фильтрам
func (r *SubscriptionRepository) GetTotalCost(ctx context.Context, filters model.CostAnalyticsRequest) (int, error) {
	baseQuery := `SELECT SUM(price) FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argCounter := 1
	if filters.UserID != "" {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argCounter)
		args = append(args, filters.UserID)
		argCounter++
	}
	if filters.ServiceName != "" {
		baseQuery += fmt.Sprintf(" AND service_name = $%d", argCounter)
		args = append(args, filters.ServiceName)
		argCounter++
	}
	if filters.StartDateStr != "" && filters.EndDateStr != "" {
		baseQuery += fmt.Sprintf(" AND start_date BETWEEN $%d AND $%d", argCounter, argCounter+1)
		args = append(args, filters.StartDateStr)
		args = append(args, filters.EndDateStr)
		argCounter += 2
	} else if filters.StartDateStr != "" {
		baseQuery += fmt.Sprintf(" AND start_date >= $%d", argCounter)
		args = append(args, filters.StartDateStr)
		argCounter++
	} else if filters.EndDateStr != "" {
		baseQuery += fmt.Sprintf(" AND start_date <= $%d", argCounter)
		args = append(args, filters.EndDateStr)
		argCounter++
	}
	var totalCost sql.NullInt64
	err := r.DB.QueryRowContext(ctx, baseQuery, args...).Scan(&totalCost)
	if err != nil {
		log.Printf("ERROR: Failed to execute GetTotalCost analytics query: %v", err)
		return 0, fmt.Errorf("error when executing an analytics request: %w", err)
	}
	if !totalCost.Valid {
		return 0, nil
	}
	return int(totalCost.Int64), nil
}
