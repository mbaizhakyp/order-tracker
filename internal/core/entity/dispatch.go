package entity

import (
	"time"

	"github.com/google/uuid"
)

type OfferStatus string

const (
	OfferStatusPending  OfferStatus = "PENDING"
	OfferStatusAccepted OfferStatus = "ACCEPTED"
	OfferStatusRejected OfferStatus = "REJECTED"
	OfferStatusExpired  OfferStatus = "EXPIRED"
)

type DispatchOffer struct {
	ID        uuid.UUID   `json:"id"`
	OrderID   uuid.UUID   `json:"order_id"`
	ShopperID uuid.UUID   `json:"shopper_id"`
	Status    OfferStatus `json:"status"`
	ExpiresAt time.Time   `json:"expires_at"`
}
