package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated           OrderStatus = "CREATED"
	OrderStatusOffered           OrderStatus = "OFFERED"
	OrderStatusClaimed           OrderStatus = "CLAIMED"
	OrderStatusArrivedAtStore    OrderStatus = "ARRIVED_AT_STORE"
	OrderStatusPickedUp          OrderStatus = "PICKED_UP"
	OrderStatusArrivedAtCustomer OrderStatus = "ARRIVED_AT_CUSTOMER"
	OrderStatusDelivered         OrderStatus = "DELIVERED"
	OrderStatusCancelled         OrderStatus = "CANCELLED"
)

type OrderEvent struct {
	ID        uuid.UUID       `json:"id"`
	OrderID   uuid.UUID       `json:"order_id"`
	Status    string          `json:"status"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

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
	DeliveryLat float64         `json:"delivery_lat"`
	DeliveryLng float64         `json:"delivery_lng"`
	CreatedAt   time.Time       `json:"created_at"`
	StoreName   string          `json:"store_name,omitempty"`   // Populated in list views
	ShopperName string          `json:"shopper_name,omitempty"` // Populated in details views
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
