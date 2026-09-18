package models

import "time"

type Event struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	TotalSlots     int       `json:"total_slots"`
	AvailableSlots int       `json:"available_slots"`
	CreatedAt      time.Time `json:"created_at"`
}
