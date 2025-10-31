package models

import "time"

type CartItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type Cart struct {
	ShoppingCartID int64      `json:"shopping_cart_id"`
	CustomerID     int64      `json:"customer_id"`
	Status         string     `json:"status"`
	Items          []CartItem `json:"items"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
