package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Config
const (
	ServerURL      = "http://localhost:8080/api/v1/location"
	ShopperCount   = 5
	UpdateInterval = 2 * time.Second
	CenterLat      = 30.2672 // Austin, TX (Target location)
	CenterLng      = -97.7431
	MovementJitter = 0.0005 // Approx 50 meters
)

type Shopper struct {
	ID  uuid.UUID
	Lat float64
	Lng float64
}

func main() {
	var shoppers []*Shopper

	// 1. Initialize Shoppers around the center
	// 1. Initialize Shoppers
	// Fixed list of IDs to ensure they exist in the DB
	shopperIDs := []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a34",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a35",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a36",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a37",
	}

	for i := 0; i < ShopperCount; i++ {
		id := uuid.MustParse(shopperIDs[i])
		shoppers = append(shoppers, &Shopper{
			ID:  id,
			Lat: CenterLat + (rand.Float64()*0.02 - 0.01),
			Lng: CenterLng + (rand.Float64()*0.02 - 0.01),
		})
	}

	fmt.Printf("Started simulation with %d shoppers around %.4f, %.4f\n", ShopperCount, CenterLat, CenterLng)
	fmt.Printf("Press Ctrl+C to stop...\n")

	client := &http.Client{Timeout: 1 * time.Second}

	// 2. Main Loop
	ticker := time.NewTicker(UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		for _, s := range shoppers {
			// Random Walk
			s.Lat += (rand.Float64()*MovementJitter*2 - MovementJitter)
			s.Lng += (rand.Float64()*MovementJitter*2 - MovementJitter)

			// Send Update
			go sendLocation(client, s)
		}
	}
}

func sendLocation(client *http.Client, s *Shopper) {
	payload := map[string]interface{}{
		"shopper_id": s.ID,
		"lat":        s.Lat,
		"lng":        s.Lng,
	}
	body, _ := json.Marshal(payload)

	resp, err := client.Post(ServerURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to update shopper %s: %v", s.ID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error updating shopper %s: Status %d", s.ID, resp.StatusCode)
	} else {
		// Optional: Verbose logging
		// fmt.Printf("Updated %s: %.4f, %.4f\n", s.ID, s.Lat, s.Lng)
	}
}
