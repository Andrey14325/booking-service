package service

import (
	"booking-service/internal/storage"
	"context"
	"errors"
	"testing"
)

type mockReservationStorage struct {
	createReservationFunc func(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error)
}

func (m *mockReservationStorage) CreateReservation(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
	if m.createReservationFunc == nil {
		panic("CreateReservation was not expected to be called in this test")
	}
	return m.createReservationFunc(ctx, reservationID, eventID, userID, idempotencyKey)
}

func TestReservationService_CreateReservation_EmptyUserID(t *testing.T) {
	mock := &mockReservationStorage{
		createReservationFunc: func(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
			t.Fatal("storage should not be called when validation fails")
			return "", nil
		},
	}

	srv := NewReservationService(mock)
	gstr := "gstr"
	iKey := &gstr
	_, err := srv.CreateReservation(context.Background(), "event", "", iKey)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestReservationService_CreateReservation_EventNotFound(t *testing.T) {
	mock := &mockReservationStorage{
		createReservationFunc: func(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
			//t.Fatal("storage should not be called when validation fails")
			return "", storage.ErrEventNotFound
		},
	}

	srv := NewReservationService(mock)
	gstr := "gstr"
	iKey := &gstr
	_, err := srv.CreateReservation(context.Background(), "event", "user", iKey)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReservationService_CreateReservation_NoSlots(t *testing.T) {
	mock := &mockReservationStorage{
		createReservationFunc: func(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
			return "", storage.ErrNoAvailableSlots
		},
	}

	srv := NewReservationService(mock)

	gstr := "gstr"
	iKey := &gstr

	_, err := srv.CreateReservation(context.Background(), "event", "efF", iKey)

	if !errors.Is(err, ErrNoSlots) {
		t.Fatalf("expected ErrNoSlots, got %v", err)
	}
}

func TestReservationService_CreateReservation_Success(t *testing.T) {
	mock := &mockReservationStorage{
		createReservationFunc: func(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
			return "res_id", nil
		},
	}

	srv := NewReservationService(mock)

	gstr := "gstr"
	iKey := &gstr

	id, err := srv.CreateReservation(context.Background(), "event", "efF", iKey)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id == "" {
		t.Fatal("expected non-empty event id")
	}
}

func TestReservationService_CreateReservation_NilIdempotencyKey(t *testing.T) {
	mock := &mockReservationStorage{
		createReservationFunc: func(ctx context.Context, reservationID, eventID, userID string, idempotencyKey *string) (string, error) {
			if idempotencyKey != nil {
				t.Errorf("expected nil idempotency key, got %v", *idempotencyKey)
			}
			return "res_id", nil
		},
	}

	srv := NewReservationService(mock)
	id, err := srv.CreateReservation(context.Background(), "event", "user", nil)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id != "res_id" {
		t.Errorf("expected id 'res_id', got %s", id)
	}
}
