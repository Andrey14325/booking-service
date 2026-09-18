package service

import (
	"booking-service/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrValidation = errors.New("validation error")
var ErrNotFound = errors.New("not found")

type EventStorage interface {
	CreateEvent(ctx context.Context, id, name string, totalSlots int) error
	GetEvent(ctx context.Context, id string) (models.Event, error)
}

type EventService struct {
	storage EventStorage
}

func NewEventService(storage EventStorage) *EventService {
	return &EventService{storage: storage}
}

func (s *EventService) CreateEvent(ctx context.Context, name string, totalSlots int) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name is required: %w", ErrValidation)
	}

	if totalSlots <= 0 {
		return "", fmt.Errorf("total slots must be positive: %w", ErrValidation)
	}

	id := uuid.NewString()
	if err := s.storage.CreateEvent(ctx, id, name, totalSlots); err != nil {
		return "", fmt.Errorf("create event: %w", err)
	}

	return id, nil
}

func (s *EventService) GetEvent(ctx context.Context, id string) (models.Event, error) {
	evAns, err := s.storage.GetEvent(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Event{}, fmt.Errorf("get event: %w", ErrNotFound)
		}
		return models.Event{}, fmt.Errorf("get event: %w", err)
	}

	return evAns, nil
}
