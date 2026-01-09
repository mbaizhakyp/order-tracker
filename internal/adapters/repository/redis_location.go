package repository

import (
	"context"
	"fmt"

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
