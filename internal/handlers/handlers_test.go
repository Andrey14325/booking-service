package handlers

import (
	"booking-service/internal/models"
	"booking-service/internal/service"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp["id"] != "fake_id" {
		t.Errorf("expected id 'fake_id', got %s", resp["id"])
	}
}

func TestHandler_CreateEvent_InvalidJSON(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			CreateEventFunc: func(ctx context.Context, name string, totalSlots int) (string, error) {
				t.Fatal("service should not be called when validation fails")
				return "", nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected StatusBadRequest-400, got %v", rec.Code)
	}
}

func TestHandler_CreateEvent_ValidationError(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			CreateEventFunc: func(ctx context.Context, name string, totalSlots int) (string, error) {
				return "", service.ErrValidation
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"name":"","total_slots":-10}`))
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected StatusBadRequest-400, got %v", rec.Code)
	}
}

func TestHandler_CreateEvent_InternalError(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			CreateEventFunc: func(ctx context.Context, name string, totalSlots int) (string, error) {
				return "", errors.New("some error")
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"name":"Concert","total_slots":10}`))
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected StatusInternalServerError-500, got %v", rec.Code)
	}
}

func TestHandler_GetEvent_Success(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			GetEventFunc: func(ctx context.Context, id string) (models.Event, error) {
				return models.Event{
					ID:             "00000000-0000-0000-0000-000000000000",
					Name:           "some_name",
					TotalSlots:     10,
					AvailableSlots: 100,
					CreatedAt:      time.Now(),
				}, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/00000000-0000-0000-0000-000000000000", nil)
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
	rec := httptest.NewRecorder()

	h.GetEvent(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected StatusOK-200, got %v", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var resp models.Event
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.ID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("expected id '00000000-0000-0000-0000-000000000000', got %s", resp.ID)
	}
}

func TestHandler_GetEvent_NotFound(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			GetEventFunc: func(ctx context.Context, id string) (models.Event, error) {
				return models.Event{}, service.ErrNotFound
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/00000000-0000-0000-0000-000000000000", nil)
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
	rec := httptest.NewRecorder()

	h.GetEvent(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected StatusNotFound-404, got %v", rec.Code)
	}
}

func TestHandler_GetEvent_InvalidUUID(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		eventService: &mockEventService{
			GetEventFunc: func(ctx context.Context, id string) (models.Event, error) {
				t.Fatal("service should not be called when validation fails")
				return models.Event{}, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/bad_id", nil)
	req.SetPathValue("id", "bad_id")
	rec := httptest.NewRecorder()

	h.GetEvent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected StatusBadRequest-400, got %v", rec.Code)
	}
}

func TestHandler_CreateReservation_Success(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		reservationService: &mockReservationService{
			CreateReservationFunc: func(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error) {
				return "some_id", nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/00000000-0000-0000-0000-000000000000/reservations", strings.NewReader(`{"user_id":"user_123"}`))
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
	req.Header.Add("Idempotency-Key", "some_key")
	rec := httptest.NewRecorder()

	h.CreateReservation(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected StatusCreated-201, got %v", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if resp["id"] != "some_id" {
		t.Errorf("expected id 'some_id', got %s", resp["id"])
	}
}

func TestHandler_CreateReservation_EventNotFound(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		reservationService: &mockReservationService{
			CreateReservationFunc: func(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error) {
				return "", service.ErrNotFound
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/00000000-0000-0000-0000-000000000000/reservations", strings.NewReader(`{"user_id":"user_123"}`))
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
	req.Header.Add("Idempotency-Key", "some_key")
	rec := httptest.NewRecorder()

	h.CreateReservation(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected StatusNotFound-404, got %v", rec.Code)
	}
}

func TestHandler_CreateReservation_NoSlots(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		reservationService: &mockReservationService{
			CreateReservationFunc: func(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error) {
				return "", service.ErrNoSlots
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/00000000-0000-0000-0000-000000000000/reservations", strings.NewReader(`{"user_id":"user_123"}`))
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
	req.Header.Add("Idempotency-Key", "some_key")
	rec := httptest.NewRecorder()

	h.CreateReservation(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected StatusConflict-409, got %v", rec.Code)
	}
}

func TestHandler_CreateReservation_ValidationError(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		reservationService: &mockReservationService{
			CreateReservationFunc: func(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error) {
				return "", service.ErrValidation
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/00000000-0000-0000-0000-000000000000/reservations", strings.NewReader(`{"user_id":""}`))
	req.SetPathValue("id", "00000000-0000-0000-0000-000000000000")
	req.Header.Add("Idempotency-Key", "some_key")
	rec := httptest.NewRecorder()

	h.CreateReservation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected StatusBadRequest-400, got %v", rec.Code)
	}
}

func TestHandler_CreateReservation_InvalidEventID(t *testing.T) {
	h := &Handler{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		reservationService: &mockReservationService{
			CreateReservationFunc: func(ctx context.Context, eventID, userID string, idempotencyKey *string) (string, error) {
				t.Fatal("service should not be called when validation fails")
				return "", nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/bad_id/reservations", strings.NewReader(`{"user_id":"user_123"}`))
	req.SetPathValue("id", "bad_id")
	req.Header.Add("Idempotency-Key", "some_key")
	rec := httptest.NewRecorder()

	h.CreateReservation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected StatusBadRequest-400, got %v", rec.Code)
	}
}
