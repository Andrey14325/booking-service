package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoAvailableSlots = errors.New("no available slots")
var ErrEventNotFound = errors.New("event not found")

type ReservationStorage struct {
	pool *pgxpool.Pool
}

func NewReservationStorage(pool *pgxpool.Pool) *ReservationStorage {
	return &ReservationStorage{pool: pool}
}

func (s *ReservationStorage) CreateReservation(ctx context.Context, reservationID, eventID, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var availableSlots int
	err = tx.QueryRow(ctx,
		`SELECT available_slots FROM events WHERE events_id = $1 FOR UPDATE`,
		eventID,
	).Scan(&availableSlots)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("get event: %w", ErrEventNotFound)
		}
		return fmt.Errorf("get event: %w", err)
	}

	if availableSlots <= 0 {
		return ErrNoAvailableSlots
	}

	_, err = tx.Exec(ctx, `
		UPDATE events
		SET available_slots = available_slots - 1
		WHERE events_id = $1
		`, eventID)
	if err != nil {
		return fmt.Errorf("events update: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO reservations (
		reservations_id,
		event_id,
		user_id,
		status
		)
		VALUES ($1, $2, $3, $4)`,
		reservationID, eventID, userID, "pending",
	)
	if err != nil {
		return fmt.Errorf("create reservation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
