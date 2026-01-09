package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusOffered   OrderStatus = "OFFERED"
	OrderStatusClaimed   OrderStatus = "CLAIMED"
	OrderStatusPickedUp  OrderStatus = "PICKED_UP"
	OrderStatusDelivered OrderStatus = "DELIVERED"
)

type OrderItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"qty"`
	ImageURL string `json:"image_url"`
}

type Order struct {
	ID          uuid.UUID       `json:"id"`
	CustomerID  uuid.UUID       `json:"customer_id"`
	StoreID     uuid.UUID       `json:"store_id"`
	ShopperID   *uuid.UUID      `json:"shopper_id,omitempty"` // Nullable
	Status      OrderStatus     `json:"status"`
	TotalAmount int64           `json:"total_amount"` // In cents
	Items       json.RawMessage `json:"items"`        // JSONB: []OrderItem
	CreatedAt   time.Time       `json:"created_at"`
}

// Helper to parse items
func (o *Order) GetItems() ([]OrderItem, error) {
	var items []OrderItem
	if len(o.Items) == 0 {
		return items, nil
	}
	err := json.Unmarshal(o.Items, &items)
	return items, err
}

func (o *Order) SetItems(items []OrderItem) error {
	bytes, err := json.Marshal(items)
	if err != nil {
		return err
	}
	o.Items = bytes
	return nil
}
