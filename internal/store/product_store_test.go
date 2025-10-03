package store

import (
	"CS6650_Online_Store/internal/models"
	"testing"
)

func TestProductStore_AddOrUpdateProduct(t *testing.T) {
	store := NewProductStore()

	product := &models.Product{
		ProductID:    1,
		SKU:          "ABC123",
		Manufacturer: "Test Manufacturer",
		CategoryID:   1,
		Weight:       100,
		SomeOtherID:  1,
	}

	// Test adding a new product
	err := store.AddOrUpdateProduct(product)
	if err != nil {
		t.Errorf("AddOrUpdateProduct() error = %v", err)
	}

	// Verify product was added
	if !store.ProductExists(1) {
		t.Error("Product should exist after adding")
	}

	// Test updating existing product
	product.SKU = "XYZ789"
	err = store.AddOrUpdateProduct(product)
	if err != nil {
		t.Errorf("AddOrUpdateProduct() error = %v", err)
	}

	// Verify product was updated
	retrievedProduct, err := store.GetProduct(1)
	if err != nil {
		t.Errorf("GetProduct() error = %v", err)
	}
	if retrievedProduct.SKU != "XYZ789" {
		t.Errorf("Expected SKU to be updated to XYZ789, got %s", retrievedProduct.SKU)
	}
}

func TestProductStore_GetProduct(t *testing.T) {
	store := NewProductStore()

	product := &models.Product{
		ProductID:    1,
		SKU:          "ABC123",
		Manufacturer: "Test Manufacturer",
		CategoryID:   1,
		Weight:       100,
		SomeOtherID:  1,
	}

	// Test getting non-existent product
	_, err := store.GetProduct(1)
	if err != ErrProductNotFound {
		t.Errorf("Expected ErrProductNotFound, got %v", err)
	}

	// Add product and test retrieval
	store.AddOrUpdateProduct(product)

	retrievedProduct, err := store.GetProduct(1)
	if err != nil {
		t.Errorf("GetProduct() error = %v", err)
	}

	// Verify it's a copy (different memory addresses)
	if retrievedProduct == product {
		t.Error("GetProduct should return a copy, not the original pointer")
	}

	// Verify content is the same
	if retrievedProduct.SKU != product.SKU {
		t.Errorf("Expected SKU %s, got %s", product.SKU, retrievedProduct.SKU)
	}
}

func TestProductStore_ProductExists(t *testing.T) {
	store := NewProductStore()

	// Test non-existent product
	if store.ProductExists(1) {
		t.Error("Product should not exist initially")
	}

	// Add product
	product := &models.Product{
		ProductID:    1,
		SKU:          "ABC123",
		Manufacturer: "Test Manufacturer",
		CategoryID:   1,
		Weight:       100,
		SomeOtherID:  1,
	}
	store.AddOrUpdateProduct(product)

	// Test existing product
	if !store.ProductExists(1) {
		t.Error("Product should exist after adding")
	}
}

func TestProductStore_GetAllProducts(t *testing.T) {
	store := NewProductStore()

	// Test empty store
	products := store.GetAllProducts()
	if len(products) != 0 {
		t.Errorf("Expected 0 products, got %d", len(products))
	}

	// Add multiple products
	product1 := &models.Product{ProductID: 1, SKU: "ABC123", Manufacturer: "Mfg1", CategoryID: 1, Weight: 100, SomeOtherID: 1}
	product2 := &models.Product{ProductID: 2, SKU: "DEF456", Manufacturer: "Mfg2", CategoryID: 2, Weight: 200, SomeOtherID: 2}

	store.AddOrUpdateProduct(product1)
	store.AddOrUpdateProduct(product2)

	// Test retrieval
	products = store.GetAllProducts()
	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}

	// Verify they are copies
	for _, p := range products {
		if p == product1 || p == product2 {
			t.Error("GetAllProducts should return copies, not original pointers")
		}
	}
}

// Test concurrent access (basic test)
func TestProductStore_ConcurrentAccess(t *testing.T) {
	store := NewProductStore()

	// This is a basic test - in real scenarios you'd want more sophisticated concurrency testing
	done := make(chan bool, 2)

	// Goroutine 1: Add products
	go func() {
		for i := 1; i <= 10; i++ {
			product := &models.Product{
				ProductID:    int32(i),
				SKU:          "ABC" + string(rune(i)),
				Manufacturer: "Test",
				CategoryID:   1,
				Weight:       100,
				SomeOtherID:  1,
			}
			store.AddOrUpdateProduct(product)
		}
		done <- true
	}()

	// Goroutine 2: Read products
	go func() {
		for i := 1; i <= 10; i++ {
			store.ProductExists(int32(i))
			store.GetProduct(int32(i)) // This might error, but shouldn't panic
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify final state
	products := store.GetAllProducts()
	if len(products) != 10 {
		t.Errorf("Expected 10 products after concurrent operations, got %d", len(products))
	}
}
