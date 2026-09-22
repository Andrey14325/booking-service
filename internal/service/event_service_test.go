package service

import (
	"booking-service/internal/models"
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type mockEventStorage struct {
	createEventFunc func(ctx context.Context, id, name string, totalSlots int) error
	getEventFunc    func(ctx context.Context, id string) (models.Event, error)
}

func (m *mockEventStorage) CreateEvent(ctx context.Context, id, name string, totalSlots int) error {
	if m.createEventFunc == nil {
		panic("CreateEvent was not expected to be called in this test")
	}
	return m.createEventFunc(ctx, id, name, totalSlots)
}

func (m *mockEventStorage) GetEvent(ctx context.Context, id string) (models.Event, error) {
	if m.getEventFunc == nil {
		panic("GetEvent was not expected to be called in this test")
	}
	return m.getEventFunc(ctx, id)
}

func TestEventService_CreateEvent_EmptyName(t *testing.T) {
	mock := &mockEventStorage{
		createEventFunc: func(ctx context.Context, id, name string, totalSlots int) error {
			t.Fatal("storage should not be called when validation fails")
			return nil
		},
	}

	svc := NewEventService(mock)

	_, err := svc.CreateEvent(context.Background(), "", 10)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestEventService_CreateEvent_NegativeSlots(t *testing.T) {
	mock := &mockEventStorage{
		createEventFunc: func(ctx context.Context, id, name string, totalSlots int) error {
			t.Fatal("storage should not be called when validation fails")
			return nil
		},
	}

	svc := NewEventService(mock)

	_, err := svc.CreateEvent(context.Background(), "", -10)

	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got %v", err)
	}
}

func TestEventService_CreateEvent_Success(t *testing.T) {
	mock := &mockEventStorage{
		createEventFunc: func(ctx context.Context, id, name string, totalSlots int) error {
			return nil
		},
	}

	svc := NewEventService(mock)

	id, err := svc.CreateEvent(context.Background(), "Concert", 10)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if id == "" {
		t.Errorf("expected non-empty id, got empty string")
	}
}

func TestEventService_GetEvent_NotFound(t *testing.T) {
	mock := &mockEventStorage{
		getEventFunc: func(ctx context.Context, id string) (models.Event, error) {
			return models.Event{}, pgx.ErrNoRows
		},
	}
	srv := NewEventService(mock)
	_, err := srv.GetEvent(context.Background(), "gsdraf")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestEventService_GetEvent_Success(t *testing.T) {
	mock := &mockEventStorage{
		getEventFunc: func(ctx context.Context, id string) (models.Event, error) {
			return models.Event{ID: "evt-123", Name: "Concert"}, nil
		},
	}
	srv := NewEventService(mock)
	mod, err := srv.GetEvent(context.Background(), "evt-123")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if mod.ID == "" {
		t.Fatal("expected non-empty event id")
	}
}
