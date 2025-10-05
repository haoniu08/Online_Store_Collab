package main

import (
	"CS6650_Online_Store/internal/models"
	"CS6650_Online_Store/internal/store"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Testing Product and ProductStore...")

	// Test 1: Create and validate a product
	fmt.Println("\n=== Test 1: Product Validation ===")
	product := &models.Product{
		ProductID:    1,
		SKU:          "ABC123",
		Manufacturer: "Test Manufacturer",
		CategoryID:   1,
		Weight:       100,
		SomeOtherID:  1,
	}

	if err := product.Validate(); err != nil {
		log.Printf("Product validation failed: %v", err)
	} else {
		fmt.Println("✓ Product validation passed")
	}

	// Test 2: Test invalid product
	fmt.Println("\n=== Test 2: Invalid Product ===")
	invalidProduct := &models.Product{
		ProductID:    0, // Invalid: must be >= 1
		SKU:          "",
		Manufacturer: "Test",
		CategoryID:   1,
		Weight:       -5, // Invalid: must be >= 0
		SomeOtherID:  1,
	}

	if err := invalidProduct.Validate(); err != nil {
		fmt.Printf("✓ Invalid product correctly rejected: %v\n", err)
	} else {
		fmt.Println("✗ Invalid product was accepted (this is wrong!)")
	}

	// Test 3: ProductStore operations
	fmt.Println("\n=== Test 3: ProductStore Operations ===")
	productStore := store.NewProductStore()

	// Test adding product
	err := productStore.AddOrUpdateProduct(product)
	if err != nil {
		log.Printf("Failed to add product: %v", err)
		return
	}
	fmt.Println("✓ Product added successfully")

	// Test checking if product exists
	if productStore.ProductExists(1) {
		fmt.Println("✓ Product exists check passed")
	} else {
		fmt.Println("✗ Product should exist but doesn't")
	}

	// Test retrieving product
	retrievedProduct, err := productStore.GetProduct(1)
	if err != nil {
		log.Printf("Failed to retrieve product: %v", err)
		return
	}
	fmt.Printf("✓ Retrieved product: ID=%d, SKU=%s\n", retrievedProduct.ProductID, retrievedProduct.SKU)

	// Test updating product
	fmt.Println("\n=== Test 4: Update Product ===")
	product.SKU = "XYZ789"
	product.Weight = 150

	err = productStore.AddOrUpdateProduct(product)
	if err != nil {
		log.Printf("Failed to update product: %v", err)
		return
	}

	updatedProduct, err := productStore.GetProduct(1)
	if err != nil {
		log.Printf("Failed to retrieve updated product: %v", err)
		return
	}
	fmt.Printf("✓ Updated product: SKU=%s, Weight=%d\n", updatedProduct.SKU, updatedProduct.Weight)

	// Test 5: Add multiple products
	fmt.Println("\n=== Test 5: Multiple Products ===")
	for i := 2; i <= 5; i++ {
		p := &models.Product{
			ProductID:    int32(i),
			SKU:          fmt.Sprintf("PROD%d", i),
			Manufacturer: "Test Manufacturer",
			CategoryID:   1,
			Weight:       int32(i * 10),
			SomeOtherID:  1,
		}
		productStore.AddOrUpdateProduct(p)
	}

	allProducts := productStore.GetAllProducts()
	fmt.Printf("✓ Total products in store: %d\n", len(allProducts))

	for _, p := range allProducts {
		fmt.Printf("  - Product ID: %d, SKU: %s, Weight: %d\n",
			p.ProductID, p.SKU, p.Weight)
	}

	// Test 6: Error handling
	fmt.Println("\n=== Test 6: Error Handling ===")
	_, err = productStore.GetProduct(999) // Non-existent product
	if err != nil {
		fmt.Printf("✓ Correctly handled non-existent product: %v\n", err)
	} else {
		fmt.Println("✗ Should have returned error for non-existent product")
	}

	fmt.Println("\n=== All Tests Completed ===")
}
