package models

import (
	"encoding/json"
	"errors"
)

// Product represents a product in the e-commerce system
type Product struct {
	// Existing fields for compatibility
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`

	// New fields for Homework 6 search functionality
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

// Error represents an API error response
type Error struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SearchResponse represents the response format for product search
type SearchResponse struct {
	Products   []Product `json:"products"`              // Max 20 results
	TotalFound int       `json:"total_found"`           // Total matches found
	SearchTime string    `json:"search_time,omitempty"` // Optional search time
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

	// New field validations for Homework 6
	// name: required for search functionality
	if len(p.Name) < 1 || len(p.Name) > 200 {
		return errors.New("name must be between 1 and 200 characters")
	}

	// category: required for search functionality
	if len(p.Category) < 1 || len(p.Category) > 100 {
		return errors.New("category must be between 1 and 100 characters")
	}

	// description: optional but if provided, should have reasonable length
	if len(p.Description) > 1000 {
		return errors.New("description must be at most 1000 characters")
	}

	// brand: required for search functionality
	if len(p.Brand) < 1 || len(p.Brand) > 100 {
		return errors.New("brand must be between 1 and 100 characters")
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
