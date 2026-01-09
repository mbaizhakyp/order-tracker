package entity

import "github.com/google/uuid"

type Store struct {
	ID   uuid.UUID
	Name string
	// For simplicity in Go, we'll parse the PostGIS point into Lat/Lng
	Lat float64
	Lng float64
}
