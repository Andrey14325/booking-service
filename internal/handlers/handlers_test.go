package handlers

import (
	"booking-service/internal/models"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockEventService struct {
	CreateEventFunc func(ctx context.Context, name string, totalSlots int) (string, error)
	GetEventFunc    func(ctx context.Context, id string) (models.Event, error)
}

func (m *mockEventService) CreateEvent(ctx context.Context, name string, totalSlots int) (string, error) {
	if m.CreateEventFunc == nil {
		panic("CreateEvent was not expected to be called in this test")
	}
	return m.CreateEventFunc(ctx, name, totalSlots)
}

func (m *mockEventService) GetEvent(ctx context.Context, id string) (models.Event, error) {
	if m.GetEventFunc == nil {
		panic("GetEvent was not expected to be called in this test")
	}
	return m.GetEventFunc(ctx, id)
}

type mockReservationService struct {
	CreateReservationFunc func(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error)
}

func (m *mockReservationService) CreateReservation(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error) {
	if m.CreateReservationFunc == nil {
		panic("CreateReservation was not expected to be called in this test")
	}
	return m.CreateReservationFunc(ctx, eventID, userID, idempotencyKey)
}

func TestHandler_Health(t *testing.T) {
	h := &Handler{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestHandler_CreateEvent_Success(t *testing.T) {

	h := &Handler{logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			CreateEventFunc: func(ctx context.Context, name string, totalSlots int) (string, error) {
				return "fake_id", nil
			},
		},
		//reservationService: &mockReservationService{},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"name":"Concert","total_slots":10}`))
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
}
