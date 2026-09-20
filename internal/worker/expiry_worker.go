package worker

import (
	"context"
	"log/slog"
	"time"
)

type ReservationExpirer interface {
	ExpireReservations(ctx context.Context) (int, error)
}

type ExpiryWorker struct {
	storage  ReservationExpirer
	logger   *slog.Logger
	interval time.Duration
}

func NewExpiryWorker(storage ReservationExpirer, logger *slog.Logger, interval time.Duration) *ExpiryWorker {
	return &ExpiryWorker{
		storage:  storage,
		logger:   logger,
		interval: interval,
	}
}

func (w *ExpiryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("expiry worker stopped")
			return
		case <-ticker.C:
			cnt, err := w.storage.ExpireReservations(ctx)
			if err != nil {
				w.logger.Error("expire reservations", slog.String("error", err.Error()))
				continue
			}
			w.logger.Info("expired reservation", slog.Int("quantity", cnt))
		}
	}
}
