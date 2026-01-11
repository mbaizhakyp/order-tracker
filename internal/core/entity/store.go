package entity

import "github.com/google/uuid"

type Store struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	// For simplicity in Go, we'll parse the PostGIS point into Lat/Lng
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
