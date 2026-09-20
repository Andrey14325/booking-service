package main

import (
	"booking-service/internal/storage"
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

func main() {
	eventID := flag.String("event", "", "event id to reserve")
	workers := flag.Int("workers", 20, "number of concurrent reservation attempts")
	dsn := flag.String("dsn", "postgres://postgres:postgres@localhost:5432/booking?sslmode=disable", "database dsn")
	flag.Parse()

	if *eventID == "" {
		log.Fatal("event id is required")
	}

	ctx := context.Background()
	pool, err := storage.NewPostgresStorage(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	reservationStorage := storage.NewReservationStorage(pool)

	var successCount int64
	var failCount int64

	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Go(func() {
			reservationID := uuid.NewString()
			userID := fmt.Sprintf("user-%d", i)

			_, err := reservationStorage.CreateReservation(ctx, reservationID, *eventID, userID, nil)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}
			atomic.AddInt64(&successCount, 1)
		})
	}

	wg.Wait()

	fmt.Printf("success: %d, failed: %d\n", successCount, failCount)
}
