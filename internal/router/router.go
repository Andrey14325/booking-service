package router

import (
	"booking-service/internal/handlers"
	"booking-service/internal/server"

	"net/http"
)

func SetupRoutes(serv *server.Server, h *handlers.Handler) {
	mux := serv.GetMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/v1/events", h.CreateEvent)
	mux.HandleFunc("GET /api/v1/events/{id}", h.GetEvent)
	mux.HandleFunc("POST /api/v1/events/{id}/reserve", h.CreateReservation)

	var handler http.Handler = mux

	serv.SetHandler(handler)
}
