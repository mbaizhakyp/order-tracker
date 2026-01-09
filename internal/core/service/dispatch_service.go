package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type DispatchService struct {
	storeRepo    ports.StoreRepository
	locRepo      ports.LocationRepository
	dispatchRepo ports.DispatchRepository
	orderRepo    ports.OrderRepository
}

func NewDispatchService(
	storeRepo ports.StoreRepository,
	locRepo ports.LocationRepository,
	dispatchRepo ports.DispatchRepository,
	orderRepo ports.OrderRepository,
) *DispatchService {
	return &DispatchService{
		storeRepo:    storeRepo,
		locRepo:      locRepo,
		dispatchRepo: dispatchRepo,
		orderRepo:    orderRepo,
	}
}

// DispatchOrder is the core logic: Find shoppers near the store and send offers.
func (s *DispatchService) DispatchOrder(ctx context.Context, orderID uuid.UUID, storeID uuid.UUID) error {
	// 1. Get Store Location
	store, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return fmt.Errorf("failed to get store: %w", err)
	}

	// 2. Find Nearby Shoppers (e.g., 15km radius)
	radiusKm := 15.0
	shoppers, err := s.locRepo.GetShoppersWithinRadius(ctx, store.Lat, store.Lng, radiusKm)
	if err != nil {
		return fmt.Errorf("failed to find shoppers: %w", err)
	}

	if len(shoppers) == 0 {
		log.Printf("No shoppers found for order %s within %.1fkm", orderID, radiusKm)
		return nil // Not an error, just no luck. Maybe retry later (out of scope for now).
	}

	// 3. Create Offers
	offerCount := 0
	for _, shopper := range shoppers {
		offer := &entity.DispatchOffer{
			OrderID:   orderID,
			ShopperID: shopper.ShopperID,
			Status:    entity.OfferStatusPending,
			ExpiresAt: time.Now().Add(60 * time.Second), // 1 min to accept
		}

		if err := s.dispatchRepo.CreateOffer(ctx, offer); err != nil {
			log.Printf("Failed to create offer for shopper %s: %v", shopper.ShopperID, err)
			continue
		}
		offerCount++
	}

	// 4. Update Order Status
	if offerCount > 0 {
		if err := s.orderRepo.UpdateSTATUS(ctx, orderID, entity.OrderStatusOffered); err != nil {
			// Non-critical, but good to know
			log.Printf("Failed to update order status to OFFERED: %v", err)
		}
		log.Printf("Dispatched order %s to %d shoppers", orderID, offerCount)
	}

	return nil
}
