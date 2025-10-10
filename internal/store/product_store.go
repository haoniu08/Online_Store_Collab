package store

import (
	"CS6650_Online_Store/internal/models"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductExists   = errors.New("product already exists")
)

// ProductStore handles in-memory storage of products using sync.Map for thread safety
type ProductStore struct {
	products sync.Map // map[int32]*models.Product
	count    int32    // track total number of products for quick access
}

// Sample data arrays for product generation
var (
	brands     = []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta", "Iota", "Kappa"}
	categories = []string{"Electronics", "Books", "Home", "Sports", "Clothing", "Beauty", "Toys", "Automotive"}
)

// NewProductStore creates a new product store and generates 100,000 products
func NewProductStore() *ProductStore {
	store := &ProductStore{}

	log.Println("Generating 100,000 products...")
	store.generateProducts()
	log.Printf("Successfully generated %d products", store.count)

	return store
}

// NewEmptyProductStore creates a new product store without pre-generating products (for testing)
func NewEmptyProductStore() *ProductStore {
	return &ProductStore{}
}

// generateProducts creates 100,000 products at startup
func (s *ProductStore) generateProducts() {
	for i := int32(1); i <= 100000; i++ {
		brandIndex := (i - 1) % int32(len(brands))
		categoryIndex := (i - 1) % int32(len(categories))

		product := models.NewProduct(
			i,
			fmt.Sprintf("Product %s %d", brands[brandIndex], i),
			categories[categoryIndex],
			brands[brandIndex],
			fmt.Sprintf("Description for product %d in %s category by %s", i, categories[categoryIndex], brands[brandIndex]),
		)

		s.products.Store(i, product)
		s.count++
	}
}

// GetProduct retrieves a product by ID
func (s *ProductStore) GetProduct(productID int32) (*models.Product, error) {
	value, exists := s.products.Load(productID)
	if !exists {
		return nil, ErrProductNotFound
	}

	product := value.(*models.Product)
	// Return a copy to prevent external modification
	productCopy := *product
	return &productCopy, nil
}

// AddOrUpdateProduct adds a new product or updates existing one
func (s *ProductStore) AddOrUpdateProduct(product *models.Product) error {
	// Store a copy to prevent external modification
	productCopy := *product

	// Check if it's a new product to update count
	_, exists := s.products.Load(product.ProductID)
	s.products.Store(product.ProductID, &productCopy)

	if !exists {
		s.count++
	}

	return nil
}

// ProductExists checks if a product exists
func (s *ProductStore) ProductExists(productID int32) bool {
	_, exists := s.products.Load(productID)
	return exists
}

// GetAllProducts returns all products (useful for debugging - only use for small datasets)
func (s *ProductStore) GetAllProducts() []*models.Product {
	products := make([]*models.Product, 0, s.count)

	s.products.Range(func(key, value interface{}) bool {
		product := value.(*models.Product)
		productCopy := *product
		products = append(products, &productCopy)
		return true
	})

	return products
}

// GetProductCount returns the total number of products
func (s *ProductStore) GetProductCount() int32 {
	return s.count
}

// SearchProducts performs a bounded search through products
// This is the key method for Homework 6 - searches exactly maxCheck products
func (s *ProductStore) SearchProducts(query string, maxCheck int, maxResults int) (*models.SearchResponse, error) {
	if maxCheck <= 0 {
		maxCheck = 100 // Default to 100 as per homework requirement
	}
	if maxResults <= 0 {
		maxResults = 20 // Default max results
	}

	results := make([]models.Product, 0, maxResults)
	checked := 0
	totalFound := 0

	// Convert query to lowercase for case-insensitive search
	queryLower := strings.ToLower(query)

	// Search through products, but only check maxCheck products
	s.products.Range(func(key, value interface{}) bool {
		if checked >= maxCheck {
			return false // Stop checking after reaching maxCheck
		}

		product := value.(*models.Product)
		checked++

		// Search in name and category (case-insensitive)
		nameMatch := strings.Contains(strings.ToLower(product.Name), queryLower)
		categoryMatch := strings.Contains(strings.ToLower(product.Category), queryLower)

		if nameMatch || categoryMatch {
			totalFound++
			// Only add to results if we haven't reached maxResults
			if len(results) < maxResults {
				results = append(results, *product)
			}
		}

		return true // Continue iteration until maxCheck is reached
	})

	return &models.SearchResponse{
		Products:   results,
		TotalFound: totalFound,
	}, nil
}
