package router

import (
	"booking-service/internal/handlers"
	"booking-service/internal/server"

	"net/http"
)

func SetupRoutes(serv *server.Server, h *handlers.Handler) {
	mux := serv.GetMux()

	mux.HandleFunc("GET /health", h.Health)

	var handler http.Handler = mux

	serv.SetHandler(handler)
}
