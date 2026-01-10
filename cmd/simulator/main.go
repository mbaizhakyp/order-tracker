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
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
	"github.com/segmentio/kafka-go"
)

// Config
const (
	ServerURL      = "http://localhost:8080/api/v1"
	ShopperCount   = 5
	UpdateInterval = 500 * time.Millisecond
	CenterLat      = 33.2098 // Tuscaloosa
	CenterLng      = -87.5692
	MovementSpeed  = 0.0010 // Increased speed for demo (approx 100m/s visually)
	ArrivalRadius  = 0.0010 // Approx 100m
)

// Shopper State
type ShopperMode string

const (
	ModeIdle              ShopperMode = "IDLE"
	ModeDrivingToStore    ShopperMode = "TO_STORE"
	ModeAtStore           ShopperMode = "AT_STORE"
	ModeDrivingToCustomer ShopperMode = "TO_CUSTOMER"
	ModeAtCustomer        ShopperMode = "AT_CUSTOMER"
)

type Shopper struct {
	ID            uuid.UUID
	Lat           float64
	Lng           float64
	Mode          ShopperMode
	ActiveOrderID uuid.UUID
	TargetLat     float64
	TargetLng     float64
	mu            sync.Mutex
}

var (
	httpClient *http.Client
	shoppers   []*Shopper
	shoppersMu sync.RWMutex
	storeRepo  ports.StoreRepository
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Setup DB & Store Repo
	dbPool, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	storeRepo = repository.NewPostgresStoreRepository(dbPool)

	// 3. Initialize Shoppers
	initShoppers()

	// 4. Start Kafka Consumer (Listen for Order Events)
	go startKafkaConsumer(cfg.Kafka.Brokers)

	httpClient = &http.Client{Timeout: 1 * time.Second}

	log.Println("Simulator Started. Waiting for order events...")

	// 5. Main Simulation Loop
	ticker := time.NewTicker(UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		updateShoppers()
	}
}

func initShoppers() {
	fixedIDs := []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a34",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a35",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a36",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a37",
	}

	shoppersMu.Lock()
	defer shoppersMu.Unlock()

	for _, idStr := range fixedIDs {
		shoppers = append(shoppers, &Shopper{
			ID:   uuid.MustParse(idStr),
			Lat:  CenterLat + (rand.Float64()-0.5)*0.01,
			Lng:  CenterLng + (rand.Float64()-0.5)*0.01,
			Mode: ModeIdle,
		})
	}
}

func startKafkaConsumer(brokers []string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "orders.lifecycle", // Ensure this topic matches where orders events are published
		GroupID: "simulator-group",
	})
	defer r.Close()

	log.Println("Kafka Consumer listening on orders.lifecycle")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			time.Sleep(5 * time.Second) // Retry delay
			continue
		}
		handleMessage(m)
	}
}

type Event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func handleMessage(m kafka.Message) {
	var event Event
	if err := json.Unmarshal(m.Value, &event); err != nil {
		return
	}

	data := event.Data
	shopperIDStr, _ := data["shopper_id"].(string)
	if shopperIDStr == "" {
		return
	}
	shopperID, err := uuid.Parse(shopperIDStr)
	if err != nil {
		return
	}

	orderIDStr, _ := data["id"].(string)
	orderID, _ := uuid.Parse(orderIDStr)

	shoppersMu.Lock()
	var shopper *Shopper
	for _, s := range shoppers {
		if s.ID == shopperID {
			shopper = s
			break
		}
	}
	shoppersMu.Unlock()

	if shopper == nil {
		return
	}

	shopper.mu.Lock()
	defer shopper.mu.Unlock()

	switch event.Type {
	case "ORDER_CLAIMED":
		log.Printf("Shopper %s claimed order %s. Driving to Store.", shopper.ID, orderID)

		storeIDStr, ok := data["store_id"].(string)
		if !ok {
			log.Println("Missing store_id in payload")
			return
		}

		store, err := storeRepo.GetByID(context.Background(), uuid.MustParse(storeIDStr))
		if err == nil {
			shopper.ActiveOrderID = orderID
			shopper.Mode = ModeDrivingToStore
			shopper.TargetLat = store.Lat
			shopper.TargetLng = store.Lng
		} else {
			log.Printf("Error fetching store: %v", err)
		}

	case "ORDER_PICKED_UP":
		log.Printf("Shopper %s picked up order. Driving to Customer.", shopper.ID)
		shopper.Mode = ModeDrivingToCustomer
		lat, _ := data["delivery_lat"].(float64)
		lng, _ := data["delivery_lng"].(float64)
		shopper.TargetLat = lat
		shopper.TargetLng = lng

	case "ORDER_DELIVERED":
		log.Printf("Shopper %s delivered order. Returning to Idle.", shopper.ID)
		shopper.Mode = ModeIdle
		shopper.ActiveOrderID = uuid.Nil
	}
}

func updateShoppers() {
	shoppersMu.RLock()
	defer shoppersMu.RUnlock()

	var wg sync.WaitGroup
	for _, s := range shoppers {
		wg.Add(1)
		go func(shopper *Shopper) {
			defer wg.Done()
			shopper.mu.Lock()
			defer shopper.mu.Unlock()

			switch shopper.Mode {
			case ModeIdle:
				// Random Walk
				moveTowards(shopper, shopper.Lat+(rand.Float64()-0.5)*0.01, shopper.Lng+(rand.Float64()-0.5)*0.01, 0.0001)

			case ModeDrivingToStore:
				dist := distance(shopper.Lat, shopper.Lng, shopper.TargetLat, shopper.TargetLng)
				if dist < ArrivalRadius {
					log.Printf("Shopper %s arrived at store.", shopper.ID)
					shopper.Mode = ModeAtStore
					go postAction(shopper.ActiveOrderID, "arrive")
				} else {
					moveTowards(shopper, shopper.TargetLat, shopper.TargetLng, MovementSpeed)
				}

			case ModeDrivingToCustomer:
				dist := distance(shopper.Lat, shopper.Lng, shopper.TargetLat, shopper.TargetLng)
				if dist < ArrivalRadius {
					log.Printf("Shopper %s arrived at customer.", shopper.ID)
					shopper.Mode = ModeAtCustomer
					go postAction(shopper.ActiveOrderID, "arrive_customer")
				} else {
					moveTowards(shopper, shopper.TargetLat, shopper.TargetLng, MovementSpeed)
				}

			case ModeAtStore, ModeAtCustomer:
				// Stop and wait for state transition event
			}

			// Broadcast Location
			postLocation(shopper)

		}(s)
	}
	wg.Wait()
}

func moveTowards(s *Shopper, targetLat, targetLng, speed float64) {
	dLat := targetLat - s.Lat
	dLng := targetLng - s.Lng
	dist := math.Sqrt(dLat*dLat + dLng*dLng)

	if dist <= speed {
		s.Lat = targetLat
		s.Lng = targetLng
	} else {
		ratio := speed / dist
		s.Lat += dLat * ratio
		s.Lng += dLng * ratio
	}
}

func distance(lat1, lng1, lat2, lng2 float64) float64 {
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lng1-lng2, 2))
}

func postLocation(s *Shopper) {
	payload := map[string]interface{}{
		"shopper_id": s.ID,
		"lat":        s.Lat,
		"lng":        s.Lng,
	}
	body, _ := json.Marshal(payload)
	httpClient.Post(ServerURL+"/location", "application/json", bytes.NewBuffer(body))
}

func postAction(orderID uuid.UUID, action string) {
	if orderID == uuid.Nil {
		return
	}
	url := fmt.Sprintf("%s/orders/%s/%s", ServerURL, orderID.String(), action)
	resp, err := httpClient.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Failed to POST action %s: %v", action, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Action %s failed with status: %d", action, resp.StatusCode)
	} else {
		log.Printf("Action %s successful for order %s", action, orderID)
	}
}
