package models

import (
	"encoding/json"
	"errors"
)

// Product represents a product in the e-commerce system
type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
}

// Error represents an API error response
type Error struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Validate checks if the product data is valid according to OpenAPI spec
func (p *Product) Validate() error {
	// product_id: minimum 1
	if p.ProductID < 1 {
		return errors.New("product_id must be at least 1")
	}

	// sku: minLength 1, maxLength 100
	if len(p.SKU) < 1 || len(p.SKU) > 100 {
		return errors.New("sku must be between 1 and 100 characters")
	}

	// manufacturer: minLength 1, maxLength 200
	if len(p.Manufacturer) < 1 || len(p.Manufacturer) > 200 {
		return errors.New("manufacturer must be between 1 and 200 characters")
	}

	// category_id: minimum 1
	if p.CategoryID < 1 {
		return errors.New("category_id must be at least 1")
	}

	// weight: minimum 0
	if p.Weight < 0 {
		return errors.New("weight must be at least 0")
	}

	// some_other_id: minimum 1
	if p.SomeOtherID < 1 {
		return errors.New("some_other_id must be at least 1")
	}

	return nil
}

// ValidateProductID validates if a product ID is valid
func ValidateProductID(id int32) error {
	if id < 1 {
		return errors.New("product_id must be a positive integer")
	}
	return nil
}

// NewError creates a new Error response
func NewError(errorCode, message, details string) *Error {
	return &Error{
		Error:   errorCode,
		Message: message,
		Details: details,
	}
}

// ToJSON converts Error to JSON bytes
func (e *Error) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}
