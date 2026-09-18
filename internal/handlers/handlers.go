package handlers

import (
	"booking-service/internal/service"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

const (
	maxUploadBodySize = 1 << 20
)

type Handler struct {
	logger       *slog.Logger
	eventService *service.EventService
}

func New(logger *slog.Logger, eventService *service.EventService) *Handler {
	return &Handler{
		logger:       logger,
		eventService: eventService,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		h.logger.Error("failed to write health response", slog.String("error", err.Error()))
	}
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBodySize)
	defer r.Body.Close()

	var req struct {
		Name       string `json:"name"`
		TotalSlots int    `json:"total_slots"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		h.logger.Error("invalid create event request", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, "invalid create event request")
		return
	}

	id, err := h.eventService.CreateEvent(r.Context(), req.Name, req.TotalSlots)
	if err != nil {
		h.logger.Error("error creating event", slog.String("error", err.Error()))
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "error creating event")
			return
		}
		writeError(w, http.StatusInternalServerError, "error creating event")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"id": id}); err != nil {
		h.logger.Error("failed to write create event response", slog.String("error", err.Error()))
	}
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.logger.Error("invalid event id",
			slog.String("id", r.PathValue("id")),
			slog.String("error", err.Error()),
		)
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.eventService.GetEvent(r.Context(), id.String())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get event")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		h.logger.Error("failed to write get event response", slog.String("error", err.Error()))
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	}); err != nil {
		slog.Error("failed to send JSON response", slog.String("error", err.Error()))
	}
}
