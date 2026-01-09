package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type PostgresDispatchRepository struct {
	db *pgxpool.Pool
}

func NewPostgresDispatchRepository(db *pgxpool.Pool) ports.DispatchRepository {
	return &PostgresDispatchRepository{db: db}
}

func (r *PostgresDispatchRepository) CreateOffer(ctx context.Context, offer *entity.DispatchOffer) error {
	query := `
		INSERT INTO dispatch_offers (id, order_id, shopper_id, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if offer.ID == uuid.Nil {
		offer.ID = uuid.New()
	}
	if offer.Status == "" {
		offer.Status = entity.OfferStatusPending
	}
	createdAt := time.Now()

	_, err := r.db.Exec(ctx, query,
		offer.ID,
		offer.OrderID,
		offer.ShopperID,
		offer.Status,
		offer.ExpiresAt,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create offer: %w", err)
	}
	return nil
}
