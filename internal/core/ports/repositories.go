package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	UpdateSTATUS(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error
	ClaimOrder(ctx context.Context, orderID uuid.UUID, shopperID uuid.UUID) error
}

type LocationRepository interface {
	UpdateShopperLocation(ctx context.Context, location *entity.ShopperLocation) error
	GetShoppersWithinRadius(ctx context.Context, lat, lng, radiusKm float64) ([]entity.ShopperLocation, error)
}

type DispatchRepository interface {
	CreateOffer(ctx context.Context, offer *entity.DispatchOffer) error
}

type StoreRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Store, error)
	UpdateLocation(ctx context.Context, id string, lat, lng float64) error
}
