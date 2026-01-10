package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/adapters/repository"
	"github.com/mbaizhakyp/order-tracker/internal/core/config"
)

// Config
const (
	ServerURL      = "http://localhost:8080/api/v1/location"
	ShopperCount   = 5
	UpdateInterval = 500 * time.Millisecond // Faster updates (0.5s)
	CenterLat      = 33.2098                // Tuscaloosa, AL, USA
	CenterLng      = -87.5692
	MovementJitter = 0.0008 // Faster movement (approx 80m jumps)
)

type Shopper struct {
	ID  uuid.UUID
	Lat float64
	Lng float64
}

// Global HTTP client for updateLocation
var httpClient *http.Client

func main() {
	// 1. Load Config (to access DB)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Connect to DB to get Store Location
	// We reuse the existing repository logic or just raw SQL for simplicity in simulator main
	dbPool, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	storeRepo := repository.NewPostgresStoreRepository(dbPool)

	log.Println("Simulator waiting for store location...")

	// 3. Create Shoppers (Initially with default, but we'll update loop to follow store)
	shoppers := make([]*Shopper, ShopperCount)
	for i := 0; i < ShopperCount; i++ {
		// Use the fixed IDs we seeded in migration if you want atomic claiming to work for them
		// For now random is fine as long as we fix the IDs logic or ensure they exist.
		// Actually, we should probably fetch the shoppers from DB too or use the fixed ones.
		// Let's use the hardcoded/known UUIDs from previous steps if possible,
		// OR just keep generating random ones but acknowledge they might not be in DB (which breaks claiming).
		// Best approach: Use the fixed list or similar logic.
		// Reverting to the logic found in previous step where we had fixed IDs:
		fixedIDs := []string{
			"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33",
			"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a34",
			"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a35",
			"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a36",
			"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a37",
		}

		idStr := fixedIDs[i%len(fixedIDs)]

		shoppers[i] = &Shopper{
			ID:  uuid.MustParse(idStr),
			Lat: CenterLat, // Initial, will update
			Lng: CenterLng,
		}
	}

	log.Printf("Started simulation with %d shoppers around %.4f, %.4f\n", ShopperCount, CenterLat, CenterLng)
	log.Printf("Press Ctrl+C to stop...\n")

	httpClient = &http.Client{Timeout: 1 * time.Second} // Initialize global client

	ticker := time.NewTicker(UpdateInterval)
	defer ticker.Stop()

	// Main Loop
	for range ticker.C {
		// Fetch current store location (Dynamic!)
		// In a real app, cache this. For demo, fetching every 2s is fine for Postgres.
		store, err := storeRepo.GetByID(context.Background(), uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"))
		if err != nil {
			log.Printf("Failed to fetch store location: %v", err)
			continue
		}

		// Use store location as center
		centerLat := store.Lat
		centerLng := store.Lng

		var wg sync.WaitGroup
		for _, s := range shoppers {

			wg.Add(1)
			go func(shopper *Shopper) {
				defer wg.Done()

				// Move towards center if too far, otherwise jitter
				// Simple logic: teleport near center if far, else walk
				dist := distance(shopper.Lat, shopper.Lng, centerLat, centerLng)
				if dist > 0.05 { // > ~5km away?
					// Teleport/Respawn near store
					shopper.Lat = centerLat + (rand.Float64()-0.5)*0.01
					shopper.Lng = centerLng + (rand.Float64()-0.5)*0.01
				} else {
					// Random walk
					shopper.Lat += (rand.Float64() - 0.5) * MovementJitter
					shopper.Lng += (rand.Float64() - 0.5) * MovementJitter
				}

				if err := updateLocation(shopper); err != nil {
					log.Printf("Failed to update shopper %s: %v", shopper.ID, err)
				} else {
					log.Printf("Shopper %s moved to %.4f, %.4f", shopper.ID, shopper.Lat, shopper.Lng)
				}
			}(s)
		}
		wg.Wait()
	}
}

// Distance helper
func distance(lat1, lng1, lat2, lng2 float64) float64 {
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lng1-lng2, 2))
}

func updateLocation(s *Shopper) error {
	// Reuse existing logic
	payload := map[string]interface{}{
		"shopper_id": s.ID,
		"lat":        s.Lat,
		"lng":        s.Lng,
	}
	body, _ := json.Marshal(payload)

	resp, err := httpClient.Post(ServerURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

// Removed old sendLocation to avoid dedup
