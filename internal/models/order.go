package models

import (
	"time"
)

// Item represents an item in an order
type Item struct {
	ProductID int    `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     float64 `json:"price"`
}

// Order represents an order in the e-commerce system
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending, processing, completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// OrderStatus constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
)
