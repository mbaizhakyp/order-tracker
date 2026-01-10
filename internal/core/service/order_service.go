package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbaizhakyp/order-tracker/internal/core/entity"
	"github.com/mbaizhakyp/order-tracker/internal/core/ports"
)

type OrderService struct {
	repo      ports.OrderRepository
	publisher ports.EventPublisher
}

func NewOrderService(repo ports.OrderRepository, publisher ports.EventPublisher) *OrderService {
	return &OrderService{repo: repo, publisher: publisher}
}

type CreateOrderRequest struct {
	CustomerID  uuid.UUID          `json:"customer_id"`
	StoreID     uuid.UUID          `json:"store_id"`
	TotalCents  int64              `json:"total_cents"`
	Items       []entity.OrderItem `json:"items"`
	DeliveryLat float64            `json:"delivery_lat"`
	DeliveryLng float64            `json:"delivery_lng"`
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*entity.Order, error) {
	if req.TotalCents <= 0 {
		return nil, errors.New("total amount must be positive")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("order must have items")
	}

	order := &entity.Order{
		ID:          uuid.New(),
		CustomerID:  req.CustomerID,
		StoreID:     req.StoreID,
		Status:      entity.OrderStatusCreated,
		TotalAmount: req.TotalCents,
		DeliveryLat: req.DeliveryLat,
		DeliveryLng: req.DeliveryLng,
		CreatedAt:   time.Now(),
	}

	if err := order.SetItems(req.Items); err != nil {
		return nil, fmt.Errorf("failed to process items: %w", err)
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Publish Event "order.created"
	eventPayload := map[string]interface{}{
		"type": "ORDER_CREATED",
		"data": order,
	}
	if err := s.publisher.Publish(ctx, "orders.lifecycle", order.ID.String(), eventPayload); err != nil {
		// Log error but don't fail the request (or handle based on consistency needs)
		// For MVP, logging is enough. In prod, we might want transactional outbox.
		// For now, let's just log print (since we don't have a logger injected yet, usually we'd use zap/logrus)
		fmt.Printf("WARNING: Failed to publish event: %v\n", err)
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrderService) ClaimOrder(ctx context.Context, orderID uuid.UUID, shopperID uuid.UUID) error {
	return s.repo.ClaimOrder(ctx, orderID, shopperID)
}
