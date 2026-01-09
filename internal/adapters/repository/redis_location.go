package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

const keyShopperLocations = "shopper_locations"

type RedisLocationRepository struct {
	client *redis.Client
}

func NewRedisLocationRepository(addr string) (ports.LocationRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisLocationRepository{client: client}, nil
}

func (r *RedisLocationRepository) UpdateShopperLocation(ctx context.Context, loc *entity.ShopperLocation) error {
	// GEOADD key longitude latitude member
	// Note: Redis usage longitude first, then latitude
	err := r.client.GeoAdd(ctx, keyShopperLocations, &redis.GeoLocation{
		Name:      loc.ShopperID.String(),
		Longitude: loc.Location.Lng,
		Latitude:  loc.Location.Lat,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to update location in redis: %w", err)
	}
	return nil
}

func (r *RedisLocationRepository) GetShoppersWithinRadius(ctx context.Context, lat, lng, radiusKm float64) ([]entity.ShopperLocation, error) {
	// Query Redis for shoppers within radius
	// Using GEOSEARCH (available in Redis 6.2+)
	cmd := r.client.GeoSearch(ctx, keyShopperLocations, &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     radiusKm,
		RadiusUnit: "km",
		Sort:       "ASC", // Closest first
	})

	locations, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search shoppers: %w", err)
	}

	// Parse results
	var shoppers []entity.ShopperLocation
	for _, loc := range locations {
		// Redis returns Key as the Member name
		shopperID, err := uuid.Parse(loc)
		if err != nil {
			// Skip invalid IDs, log warn if we had a logger
			continue
		}

		// To get actual coordinates, we would need GEOSEARCH with WITHCOORD
		// But for now, we just need the IDs to dispatch to.
		// If we need the coords, we can use GeoSearchLocation
		shoppers = append(shoppers, entity.ShopperLocation{
			ShopperID: shopperID,
			// Location is not strictly needed for the offer logic right now,
			// but if we wanted it, we'd change GeoSearch to GeoSearchLocation
		})
	}

	return shoppers, nil
}
