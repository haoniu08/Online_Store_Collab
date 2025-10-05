package store

import (
	"CS6650_Online_Store/internal/models"
	"errors"
	"sync"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductExists   = errors.New("product already exists")
)

// ProductStore handles in-memory storage of products
type ProductStore struct {
	mu       sync.RWMutex
	products map[int32]*models.Product
}

// NewProductStore creates a new product store
func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[int32]*models.Product),
	}
}

// GetProduct retrieves a product by ID
func (s *ProductStore) GetProduct(productID int32) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[productID]
	if !exists {
		return nil, ErrProductNotFound
	}

	// Return a copy to prevent external modification
	productCopy := *product
	return &productCopy, nil
}

// AddOrUpdateProduct adds a new product or updates existing one
func (s *ProductStore) AddOrUpdateProduct(product *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a copy to prevent external modification
	productCopy := *product
	s.products[product.ProductID] = &productCopy

	return nil
}

// ProductExists checks if a product exists
func (s *ProductStore) ProductExists(productID int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.products[productID]
	return exists
}

// GetAllProducts returns all products (useful for debugging)
func (s *ProductStore) GetAllProducts() []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]*models.Product, 0, len(s.products))
	for _, product := range s.products {
		productCopy := *product
		products = append(products, &productCopy)
	}
	return products
}
