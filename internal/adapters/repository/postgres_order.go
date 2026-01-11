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
		SELECT o.id, o.customer_id, o.store_id, o.shopper_id, o.status, o.total_amount, o.items, o.delivery_lat, o.delivery_lng, o.created_at,
		       COALESCE(u.name, '') as shopper_name
		FROM orders o
		LEFT JOIN users u ON o.shopper_id = u.id
		WHERE o.id = $1
	`
	var order entity.Order

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
		&order.ShopperName,
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

func (r *PostgresOrderRepository) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	return r.UpdateSTATUS(ctx, orderID, entity.OrderStatusCancelled)
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

func (r *PostgresOrderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]entity.Order, error) {
	query := `
		SELECT o.id, o.store_id, s.name, o.status, o.total_amount, o.items, o.created_at
		FROM orders o
		JOIN stores s ON o.store_id = s.id
		WHERE o.customer_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		// We only scan fields needed for the history list
		if err := rows.Scan(
			&o.ID,
			&o.StoreID,
			&o.StoreName,
			&o.Status,
			&o.TotalAmount,
			&o.Items,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		o.CustomerID = customerID
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *PostgresOrderRepository) CancelStaleOrders(ctx context.Context, olderThan time.Time) (int64, error) {
	// Update all orders that are NOT delivered and NOT cancelled and are OLDER than threshold
	query := `
		UPDATE orders 
		SET status = $1 
		WHERE created_at < $2 
		AND status NOT IN ($3, $4)
	`
	tag, err := r.db.Exec(ctx, query,
		entity.OrderStatusCancelled,
		olderThan,
		entity.OrderStatusDelivered,
		entity.OrderStatusCancelled,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
