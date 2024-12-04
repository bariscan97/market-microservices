package models

import (
	"github.com/google/uuid"
)

type CartItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	ImageUrl string  `json:"image_url"`
	Category string  `json:"category"`
}

type Cart struct {
	CustomerID uuid.UUID
	CartItems  map[string]*CartItem
}
