package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type LocationService struct {
	repo      ports.LocationRepository
	publisher ports.EventPublisher
}

func NewLocationService(repo ports.LocationRepository, publisher ports.EventPublisher) *LocationService {
	return &LocationService{
		repo:      repo,
		publisher: publisher,
	}
}

type UpdateLocationRequest struct {
	ShopperID uuid.UUID `json:"shopper_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
}

func (s *LocationService) UpdateLocation(ctx context.Context, req UpdateLocationRequest) error {
	// Validation
	if req.Lat < -90 || req.Lat > 90 {
		return fmt.Errorf("invalid latitude")
	}
	if req.Lng < -180 || req.Lng > 180 {
		return fmt.Errorf("invalid longitude")
	}

	loc := &entity.ShopperLocation{
		ShopperID: req.ShopperID,
		Location: entity.GeoPoint{
			Lat: req.Lat,
			Lng: req.Lng,
		},
	}

	// Update Redis
	if err := s.repo.UpdateShopperLocation(ctx, loc); err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	// Phase 4 - Publish to Kafka for history tracking & real-time updates
	eventPayload := map[string]interface{}{
		"type": "SHOPPER_MOVED",
		"data": map[string]interface{}{
			"shopper_id": req.ShopperID,
			"lat":        req.Lat,
			"lng":        req.Lng,
			"timestamp":  time.Now(),
		},
	}

	if err := s.publisher.Publish(ctx, "tracker", req.ShopperID.String(), eventPayload); err != nil {
		// Log but don't fail the request (fire and forget)
		fmt.Printf("Failed to publish location event: %v\n", err)
	}

	// Print log for demo
	fmt.Printf("[%s] Shopper %s moved to %.4f, %.4f\n",
		time.Now().Format(time.TimeOnly), req.ShopperID, req.Lat, req.Lng)

	return nil
}
