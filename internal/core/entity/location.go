package entity

import "github.com/google/uuid"

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ShopperLocation struct {
	ShopperID uuid.UUID `json:"shopper_id"`
	Location  GeoPoint  `json:"location"`
}
