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

func (s *ReservationStorage) CreateReservation(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if idempotencyKey != nil {
		var oldReservationID string
		err := tx.QueryRow(ctx, `
			SELECT reservations_id 
			FROM reservations 
			WHERE event_id = $1 AND idempotency_key = $2`,
			eventID, *idempotencyKey).Scan(&oldReservationID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
			} else {
				return "", fmt.Errorf("check idempotency: %w", err)
			}
		} else {
			return oldReservationID, nil
		}
	}
	var availableSlots int
	err = tx.QueryRow(ctx, `
		SELECT available_slots 
		FROM events 
		WHERE events_id = $1 
		FOR UPDATE`,
		eventID,
	).Scan(&availableSlots)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("get event: %w", ErrEventNotFound)
		}
		return "", fmt.Errorf("get event: %w", err)
	}

	if availableSlots <= 0 {
		return "", ErrNoAvailableSlots
	}

	_, err = tx.Exec(ctx, `
		UPDATE events
		SET available_slots = available_slots - 1
		WHERE events_id = $1
		`, eventID)
	if err != nil {
		return "", fmt.Errorf("events update: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO reservations (
		reservations_id,
		event_id,
		user_id,
		status,
		idempotency_key,
        expires_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW() + INTERVAL '10 minutes')`,
		reservationID, eventID, userID, "pending", idempotencyKey,
	)
	if err != nil {
		return "", fmt.Errorf("create reservation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return reservationID, nil
}

func (s *ReservationStorage) ExpireReservations(ctx context.Context) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		UPDATE reservations
		SET status = 'expired'
		WHERE status = 'pending' AND expires_at < NOW()
		RETURNING event_id
	`)
	if err != nil {
		return 0, fmt.Errorf("expire reservations: %w", err)
	}
	defer rows.Close()

	var eventIDs []string

	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return 0, fmt.Errorf("rows scan error: %w", err)
		}
		eventIDs = append(eventIDs, line)
	}

	if rows.Err() != nil {
		return 0, fmt.Errorf("rows scan error: %w", rows.Err())
	}

	for _, eventID := range eventIDs {
		_, err := tx.Exec(ctx, `
			UPDATE events 
			SET available_slots = available_slots + 1 WHERE events_id = $1
		`, eventID)
		if err != nil {
			return 0, fmt.Errorf("return slot: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return len(eventIDs), nil
}
