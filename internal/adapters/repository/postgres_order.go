package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) ports.OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, store_id, status, total_amount, items, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	// Ensure ID is generated if empty
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}
	if order.Status == "" {
		order.Status = entity.OrderStatusCreated
	}

	_, err := r.db.Exec(ctx, query,
		order.ID,
		order.CustomerID,
		order.StoreID,
		order.Status,
		order.TotalAmount,
		order.Items,
		order.CreatedAt,
	)
	return err
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	query := `
		SELECT id, customer_id, store_id, shopper_id, status, total_amount, items, created_at
		FROM orders
		WHERE id = $1
	`
	var order entity.Order

	// Actually pgx supports *uuid.UUID scanning natively if setup correctly,
	// but sometimes explicit scanning is safer. Let's try direct scan first.
	// Note: sql.NullString is from database/sql. pgx uses native types.
	// For nullable uuid in pgx, we can scan into *uuid.UUID.

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.StoreID,
		&order.ShopperID,
		&order.Status,
		&order.TotalAmount,
		&order.Items,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PostgresOrderRepository) UpdateSTATUS(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}
