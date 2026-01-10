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

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) ports.OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, store_id, status, total_amount, items, delivery_lat, delivery_lng, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
		order.DeliveryLat,
		order.DeliveryLng,
		order.CreatedAt,
	)
	return err
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	query := `
		SELECT id, customer_id, store_id, shopper_id, status, total_amount, items, delivery_lat, delivery_lng, created_at
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
		&order.DeliveryLat,
		&order.DeliveryLng,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PostgresOrderRepository) UpdateSTATUS(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Update Order
	queryUpdate := `UPDATE orders SET status = $1 WHERE id = $2`
	if _, err := tx.Exec(ctx, queryUpdate, status, id); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// 2. Insert Event
	queryEvent := `INSERT INTO order_events (order_id, status) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, queryEvent, id, status); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) ClaimOrder(ctx context.Context, orderID uuid.UUID, shopperID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lock the row
	var currentStatus string
	queryLock := `SELECT status FROM orders WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRow(ctx, queryLock, orderID).Scan(&currentStatus); err != nil {
		return fmt.Errorf("failed to lock order: %w", err)
	}

	// 2. Do logic check
	if currentStatus != string(entity.OrderStatusOffered) {
		return fmt.Errorf("order cannot be claimed (status: %s)", currentStatus)
	}

	// 3. Update status and shopper_id
	queryUpdate := `UPDATE orders SET status = $1, shopper_id = $2 WHERE id = $3`
	if _, err := tx.Exec(ctx, queryUpdate, entity.OrderStatusClaimed, shopperID, orderID); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	// 4. Insert Event
	queryEvent := `INSERT INTO order_events (order_id, status) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, queryEvent, orderID, entity.OrderStatusClaimed); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) GetOrderHistory(ctx context.Context, orderID uuid.UUID) ([]entity.OrderEvent, error) {
	query := `
		SELECT id, order_id, status, metadata, created_at
		FROM order_events
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []entity.OrderEvent
	for rows.Next() {
		var e entity.OrderEvent
		if err := rows.Scan(&e.ID, &e.OrderID, &e.Status, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
