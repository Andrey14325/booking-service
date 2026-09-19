package service

import (
	"booking-service/internal/storage"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrNoSlots = errors.New("no available slots")

type ReservationStorage interface {
	CreateReservation(ctx context.Context, reservationID, eventID, userID string) error
}

type ReservationService struct {
	storage ReservationStorage
}

func NewReservationService(storage ReservationStorage) *ReservationService {
	return &ReservationService{storage: storage}
}

func (s *ReservationService) CreateReservation(ctx context.Context, eventID, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user id is required: %w", ErrValidation)
	}

	id := uuid.NewString()

	if err := s.storage.CreateReservation(ctx, id, eventID, userID); err != nil {
		if errors.Is(err, storage.ErrNoAvailableSlots) {
			return "", fmt.Errorf("create reservation: %w", ErrNoSlots)
		}
		if errors.Is(err, storage.ErrEventNotFound) {
			return "", fmt.Errorf("create reservation: %w", ErrNotFound)
		}
		return "", fmt.Errorf("create reservation: %w", err)
	}

	return id, nil
}
