package main

import (
	"booking-service/internal/config"
	"booking-service/internal/handlers"
	"booking-service/internal/router"
	"booking-service/internal/server"
	"booking-service/internal/storage"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.ParseConfig()

	handl := handlers.New(logger)

	serv, err := server.NewServer(logger, cfg.ServAddress)
	if err != nil {
		logger.Error("error creating server", slog.Any("error", err))
		os.Exit(1)
	}

	router.SetupRoutes(serv, handl)
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := storage.NewPostgresStorage(ctx, cfg.Database)
	if err != nil {
		logger.Error("error database connection", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := serv.RunServer(ctx); err != nil {
		logger.Error("error running server", slog.Any("error", err))
		os.Exit(1)
	}

}
