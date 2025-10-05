package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"CS6650_Online_Store/internal/handlers"
	"CS6650_Online_Store/internal/models"
	"CS6650_Online_Store/internal/store"

	"github.com/gorilla/mux"
)

// setupTestServer creates a test server with the same routes as main
func setupTestServer() *mux.Router {
	// Initialize store and handlers (same as main.go)
	productStore := store.NewProductStore()
	productHandler := handlers.NewProductHandler(productStore)

	// Setup router (same as main.go)
	router := mux.NewRouter()
	router.HandleFunc("/products/{productId}", productHandler.GetProduct).Methods("GET")
	router.HandleFunc("/products/{productId}/details", productHandler.AddProductDetails).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Add logging middleware (same as main.go)
	router.Use(loggingMiddleware)

	return router
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestServer()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check response body
	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("Health endpoint returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Note: Go's default handler doesn't set Content-Type for plain text
	// This is acceptable behavior
}

func TestProductWorkflow(t *testing.T) {
	router := setupTestServer()

	// Test 1: Get non-existent product (should return 404)
	t.Run("Get non-existent product", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/products/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Expected status 404, got %v", status)
		}

		// Check it's a proper JSON error response
		var errorResponse models.Error
		if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
			t.Errorf("Error response should be valid JSON: %v", err)
		}
		if errorResponse.Error != "NOT_FOUND" {
			t.Errorf("Expected error code NOT_FOUND, got %s", errorResponse.Error)
		}
	})

	// Test 2: Add a product
	t.Run("Add product", func(t *testing.T) {
		product := models.Product{
			ProductID:    1,
			SKU:          "TEST123",
			Manufacturer: "Test Manufacturer",
			CategoryID:   1,
			Weight:       100,
			SomeOtherID:  1,
		}

		jsonData, _ := json.Marshal(product)
		req, err := http.NewRequest("POST", "/products/1/details", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("Expected status 204, got %v. Response: %s", status, rr.Body.String())
		}
	})

	// Test 3: Get the product we just added (should return 200)
	t.Run("Get existing product", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/products/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status 200, got %v", status)
		}

		// Verify response is valid JSON and contains expected data
		var product models.Product
		if err := json.Unmarshal(rr.Body.Bytes(), &product); err != nil {
			t.Errorf("Response should be valid JSON: %v", err)
		}

		if product.ProductID != 1 {
			t.Errorf("Expected product_id 1, got %d", product.ProductID)
		}
		if product.SKU != "TEST123" {
			t.Errorf("Expected SKU TEST123, got %s", product.SKU)
		}

		// Check content type
		expectedContentType := "application/json"
		if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
			t.Errorf("Expected content type %s, got %s", expectedContentType, ct)
		}
	})
}

func TestInvalidRoutes(t *testing.T) {
	router := setupTestServer()

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{"Root path", "GET", "/"},
		{"Products without ID", "GET", "/products"},
		{"Products with trailing slash", "GET", "/products/"},

		{"Wrong HTTP method on health", "POST", "/health"},
		{"Wrong HTTP method on get product", "POST", "/products/1"},
		{"Wrong HTTP method on add product", "GET", "/products/1/details"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Should return 404 (route not found) or 405 (method not allowed)
			if status := rr.Code; status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 404 or 405 for %s %s, got %v", tc.method, tc.path, status)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	router := setupTestServer()

	t.Run("Invalid product ID in URL", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/products/0", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid product ID, got %v", status)
		}
	})

	t.Run("Invalid JSON in POST request", func(t *testing.T) {
		invalidJSON := `{"invalid": json}`
		req, err := http.NewRequest("POST", "/products/1/details", bytes.NewBufferString(invalidJSON))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %v", status)
		}

		// Should return proper error response
		var errorResponse models.Error
		if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
			t.Errorf("Error response should be valid JSON: %v", err)
		}
	})

	t.Run("Product ID mismatch", func(t *testing.T) {
		product := models.Product{
			ProductID:    999, // Different from URL
			SKU:          "TEST123",
			Manufacturer: "Test",
			CategoryID:   1,
			Weight:       100,
			SomeOtherID:  1,
		}

		jsonData, _ := json.Marshal(product)
		req, err := http.NewRequest("POST", "/products/1/details", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status 400 for product ID mismatch, got %v", status)
		}
	})
}

func TestMiddleware(t *testing.T) {
	router := setupTestServer()

	// Test that logging middleware is applied
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// We can't easily test the logging output without capturing logs,
	// but we can ensure the request still works with middleware
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Middleware should not break request handling: got status %v", status)
	}
}

func TestConcurrentRequests(t *testing.T) {
	router := setupTestServer()

	// Test concurrent GET requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			req, _ := http.NewRequest("GET", "/health", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Concurrent request %d failed with status %v", id, status)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark test for performance
func BenchmarkHealthEndpoint(b *testing.B) {
	router := setupTestServer()
	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
}
