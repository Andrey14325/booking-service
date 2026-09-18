package storage

import (
	"booking-service/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventStorage struct {
	pool *pgxpool.Pool
}

func NewEventStorage(pool *pgxpool.Pool) *EventStorage {
	return &EventStorage{pool: pool}
}

func (s *EventStorage) CreateEvent(ctx context.Context, id, name string, totalSlots int) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO events (
		events_id, 
		name, 
		total_slots, 
		available_slots
		) 
		VALUES ($1, $2, $3, $3)`,
		id, name, totalSlots,
	)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (s *EventStorage) GetEvent(ctx context.Context, id string) (models.Event, error) {
	pgRow := s.pool.QueryRow(ctx,
		`SELECT events_id, name, total_slots, available_slots, created_at
		FROM events
		WHERE events_id = $1`,
		id)

	var ans models.Event
	err := pgRow.Scan(&ans.ID, &ans.Name, &ans.TotalSlots, &ans.AvailableSlots, &ans.CreatedAt)
	if err != nil {
		return models.Event{}, fmt.Errorf("get event: %w", err)
	}

	return ans, nil
}
